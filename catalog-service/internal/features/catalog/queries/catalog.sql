-- name: GetAllCatalogItems :many
SELECT *
    FROM items
    ORDER BY created_at DESC;

-- name: GetCatalogItemByID :one
SELECT *
    FROM items
    WHERE id = $1;

-- name: CreateCatalogItem :one
INSERT INTO items (code, name, price, stock_quantity, image_url, created_at, updated_at)
    VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
    RETURNING *;

-- name: UpdateItemQuantity :one
UPDATE items
    SET stock_quantity = $2, updated_at = NOW()
    WHERE id = $1
    RETURNING *;