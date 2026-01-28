-- name: CreateOrder :one
INSERT INTO orders (user_id, user_name, status, order_price)
    VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_item (order_id, item_id, quantity, price_at_order)
    VALUES ($1, $2, $3, $4)
    RETURNING *;

-- name: GetAllOrders :many
SELECT
    o.id,
    o.user_id,
    o.user_name,
    o.status,
    o.order_price,
    o.created_at,
    o.updated_at,
    oi.id as item_id,
    oi.item_id as item_product_id,
    oi.quantity,
    oi.price_at_order,
    i.code,
    i.name,
    i.image_url
FROM orders o
LEFT JOIN order_item oi ON o.id = oi.order_id
LEFT JOIN items i ON oi.item_id = i.id
ORDER BY o.created_at DESC;

-- name: GetOrderByID :many
SELECT
    o.id,
    o.user_id,
    o.user_name,
    o.status,
    o.order_price,
    o.created_at,
    o.updated_at,
    oi.id as item_id,
    oi.item_id as item_product_id,
    oi.quantity,
    oi.price_at_order,
    i.code,
    i.name,
    i.image_url
FROM orders o
LEFT JOIN order_item oi ON o.id = oi.order_id
LEFT JOIN items i ON oi.item_id = i.id
WHERE o.id = $1
ORDER BY oi.created_at ASC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetItemByID :one
SELECT id, code, name, price, stock_quantity, image_url
FROM items
WHERE id = $1;

-- name: UpdateItemStock :one
UPDATE items
SET stock_quantity = stock_quantity - $2, updated_at = NOW()
WHERE id = $1 AND stock_quantity >= $2
RETURNING *;
