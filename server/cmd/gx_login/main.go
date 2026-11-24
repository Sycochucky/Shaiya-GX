// Command gx_login is the Shaiya 5.4 login server: RSA/AES key
// exchange, credential check through the original
// PS_UserData.dbo.usp_Try_GameLogin_Taiwan, server list and the session
// handoff to gx_world.
package main

import (
	"context"
	"fmt"
	"github.com/Spelunkern/Shaiya-GX/internal/authwire"
	"github.com/Spelunkern/Shaiya-GX/internal/itemstack"
	"github.com/Spelunkern/Shaiya-GX/internal/transport"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/Spelunkern/Shaiya-GX/server/internal/config"
	"github.com/Spelunkern/Shaiya-GX/server/internal/controlplane"
	"github.com/Spelunkern/Shaiya-GX/server/internal/crypto"
	"github.com/Spelunkern/Shaiya-GX/server/internal/db"
	"github.com/Spelunkern/Shaiya-GX/server/internal/logx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/netx"
	"github.com/Spelunkern/Shaiya-GX/server/internal/packet"
	"github.com/Spelunkern/Shaiya-GX/server/internal/session"
)

var nextConnection atomic.Int64
var loginByIP = netx.NewRateLimiter(4096, 30, time.Minute)
var loginByAccount = netx.NewRateLimiter(8192, 10, time.Minute)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	if err := config.LoadRuntimeEnv(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	policy, err := config.Network("login")
	if err != nil {
		return err
	}
	logger, err := logx.New("gx_login")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	defer logger.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	resetFailure := db.OnStorageFailure(func(err error) { logger.Errorf("storage unavailable; stopping service: %v", err); stop() })
	defer resetFailure()
	defer db.Close()
	logger.Printf("Starting GX Server Login")
	if err := db.Open(db.UserData); err != nil {
		logger.Printf("GX Server Login database error: %v", err)
		return err
	}
	loginByIP = netx.NewRateLimiter(4096, policy.IPAttempts, time.Minute)
	loginByAccount = netx.NewRateLimiter(8192, policy.AccountAttempts, time.Minute)
	listener, err := transport.Listen(config.LoginListen())
	if err != nil {
		logger.Errorf("GX Server Login failed to bind %s: %v", config.LoginListen(), err)
		return err
	}
	logger.Printf("GX Server Login listening on %s", config.LoginListen())

	admin, err := controlplane.Start(ctx, "gx_login", os.Getenv("RAPTURE_LOGIN_DIAGNOSTICS"), stop, func() controlplane.ServiceStatus {
		return controlplane.ServiceStatus{Endpoint: listener.Addr().String()}
	}, db.Observe)
	if err != nil {
		listener.Close()
		logger.Printf("local control startup failed: %v", err)
		return err
	}
	defer admin.Close()
	logger.Event("local diagnostics ready", "address", admin.Address())
	if err := netx.Serve(ctx, listener, policy.Connections, policy.PerIP, func(conn net.Conn) {
		log := logger.Tagged(fmt.Sprintf("conn%d", nextConnection.Add(1)))
		defer func() {
			if r := recover(); r != nil {
				log.Printf("login session panic: %v", r)
			}
		}()
		if err := handle(conn, log); err != nil {
			log.Printf("login client error: %v", err)
		}
	}, netx.NewRateLimiter(4096, policy.Establishments, time.Minute)); err != nil {
		logger.Printf("server accept error: %v", err)
		return err
	}

	return nil
}

func handshake() []byte {
	exponent, modulus := crypto.LoginPublicKey()
	w := packet.New(packet.LoginHandshake)
	w.U8(0)
	w.U8(uint8(len(exponent)))
	w.U8(uint8(len(modulus)))
	w.Bytes(exponent[:])
	w.Bytes(modulus[:])
	return w.MustFinish()
}

func loginReply(result uint8, userUID uint32, permission uint8, identity [16]byte) []byte {
	w := packet.New(packet.LoginRequest)
	w.U8(result)
	if result == 0 {
		w.U32(userUID)
		w.U8(permission)
		w.Bytes(identity[:])
	}
	return w.MustFinish()
}

func serverList() []byte {
	w := packet.New(packet.ServerList)
	w.U8(1)    // server count
	w.U8(1)    // server id
	w.U8(0)    // status
	w.U16(0)   // users
	w.U16(100) // capacity
	w.FixedString("GX Server", 32)
	return w.MustFinish()
}

func selectServerReply(status uint8) []byte {
	w := packet.New(packet.SelectServer)
	w.U8(status)
	ip := net.ParseIP(config.PublicIP()).To4()
	if ip == nil {
		ip = net.IPv4(127, 0, 0, 1).To4()
	}
	w.Bytes(ip)
	return w.MustFinish()
}

type client struct {
	conn      net.Conn
	log       *logx.Tag
	send      *crypto.AESCTR
	recv      *crypto.AESCTR
	material  [32]byte
	userUID   uint32
	sessionID int64
}

func (c *client) write(frame []byte, label string) error {
	if err := c.conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	out := frame
	if c.send != nil && len(frame) > 2 {
		out = append(append([]byte{}, frame[:2]...), c.send.Transform(frame[2:])...)
	}
	c.log.Printf("GX Server Login sending %s (%d bytes)", label, len(frame))
	_, err := c.conn.Write(out)
	return err
}

func handle(conn net.Conn, log *logx.Tag) error {
	if err := conn.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
		return err
	}
	c := &client{conn: conn, log: log}
	remote, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	log.Printf("GX Server Login client connected from %s", remote)
	if err := c.write(handshake(), "0xA101 handshake"); err != nil {
		return err
	}
	var frames netx.Frames
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if n == 0 || err != nil {
			log.Printf("GX Server Login client disconnected")
			return nil
		}
		list, ok := frames.Push(buf[:n])
		for _, frame := range list {
			if err := c.frame(frame, remote); err != nil {
				return err
			}
		}
		if !ok {
			return fmt.Errorf("malformed frame")
		}
	}
}

func (c *client) frame(frame []byte, remote string) error {
	if c.recv == nil {
		opcode, _ := packet.U16(frame[2:4])
		if packet.Opcode(opcode) != packet.LoginHandshake || len(frame) < 5 {
			return nil
		}
		size := int(frame[4])
		if size == 0 || len(frame) < 5+size {
			return fmt.Errorf("short RSA block")
		}
		// The RSA plaintext and derived AES material are never logged.
		c.material = crypto.DeriveLoginAESMaterial(crypto.DecryptLoginRSA(frame[5 : 5+size]))
		var key, iv [16]byte
		copy(key[:], c.material[:16])
		copy(iv[:], c.material[16:])
		c.recv = crypto.NewAESCTR(key, iv)
		c.send = crypto.NewAESCTR(key, iv)
		c.log.Printf("GX Server Login key exchange complete")
		return nil
	}
	payload := c.recv.Transform(frame[2:])
	if len(payload) < 2 {
		return nil
	}
	opcode, _ := packet.U16(payload)
	switch packet.Opcode(opcode) {
	case packet.LoginRequest, packet.Opcode(authwire.Opcode):
		var username, secret string
		if packet.Opcode(opcode) == packet.LoginRequest {
			if len(payload) < 50 {
				return fmt.Errorf("short login request")
			}
			username = packet.FixedText(payload[2:34])
			secret = packet.FixedText(payload[34:50])
			if err := authwire.Validate(username, secret); err != nil {
				return err
			}
		} else {
			var err error
			username, secret, err = authwire.Decode(payload)
			if err != nil {
				return err
			}
		}
		if !loginByIP.Allow(remote) || !loginByAccount.Allow(strings.ToLower(username)) {
			return fmt.Errorf("login rate exceeded")
		}
		password := secret // never logged
		c.log.Printf("GX Server Login request username=%q", username)
		c.sessionID = session.NewSessionID()
		login, err := db.TryGameLogin(username, password, c.sessionID, remote)
		if err != nil {
			c.log.Printf("GX Server Login database error: %v", err)
			return c.write(loginReply(1, 0, 11, [16]byte{}), "0xA102 login failed")
		}
		result, permission := db.LoginDecision(login)
		if login.Status == -1 {
			result = 1
		} // same public failure as an unknown account
		if result != 0 {
			c.log.Printf("GX Server Login rejected username=%q status=%d result=%d", username, login.Status, result)
			return c.write(loginReply(result, 0, permission, [16]byte{}), "0xA102 login failed")
		}
		c.userUID = uint32(login.UserUID)
		c.log.SetTag(c.log.Tag() + "/" + username)
		identity := session.NewIdentity()
		if err := session.Save(session.Handoff{Identity: identity, AESMaterial: c.material,
			UserUID: c.userUID, SessionID: c.sessionID, CredentialVersion: login.CredentialVersion}); err != nil {
			return err
		}
		c.log.Printf("GX Server Login authenticated user_uid=%d permission=%d", c.userUID, permission)
		if err := c.write(loginReply(0, c.userUID, permission, identity), "0xA102 login success"); err != nil {
			return err
		}
		return c.write(serverList(), "0xA201 server list")
	case packet.SelectServer:
		if len(payload) < 7 || c.userUID == 0 {
			return nil
		}
		version, _ := packet.U32(payload[3:7])
		c.log.Printf("GX Server Login select server world_id=%d client_version=%d", payload[2], version)
		status := uint8(0)
		if version != itemstack.ClientVersion {
			status = 1
		}
		return c.write(selectServerReply(status), "0xA202 select server")
	}
	return nil
}
