-- Existing linking hammers use flat bonuses instead of legacy multipliers.
UPDATE PS_GameDefs_Items
SET ReqVg = CASE TypeID
    WHEN 49 THEN 10 WHEN 82 THEN 20
    WHEN 193 THEN 6 WHEN 194 THEN 7
    WHEN 195 THEN 1 WHEN 196 THEN 2 WHEN 197 THEN 3 WHEN 198 THEN 4 WHEN 199 THEN 5
END
WHERE Special = 36 AND (
    (Type IN (44, 100) AND TypeID BETWEEN 193 AND 199)
    OR (Type = 100 AND TypeID IN (49, 82))
);

-- Activate the two existing Cash Mall item IDs without changing product or inventory rows.
UPDATE PS_GameDefs_Items
SET Special = 58, ReqVg = 100, ItemName = 'GM Extraction Hammer'
WHERE Type = 100 AND TypeID = 191;

UPDATE PS_GameDefs_Items
SET Special = 36, ReqVg = 100, ItemName = 'GM Linking Hammer'
WHERE Type = 100 AND TypeID = 192;
