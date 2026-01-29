// Orders service for managing order operations

export interface OrderItem {
  productId: string;
  quantity: number;
}

export interface Order {
  id?: string;
  customerId?: string;
  customerName: string;
  items: OrderItem[];
  status?: 'pending' | 'processing' | 'completed';
  totalPrice?: number;
  createdAt?: string;
  downloadUrl?: string;
}

const ORDERS_API_URL = import.meta.env.VITE_ORDERS_API_URL;

export const ordersService = {
  // Fetch all orders
  async getAllOrders(): Promise<Order[]> {
    try {
      const response = await fetch(`${ORDERS_API_URL}`);
      if (!response.ok) {
        throw new Error('Failed to fetch orders');
      }
      const json = await response.json();

      // API returns { orders: [...] }
      let orders = Array.isArray(json.orders) ? json.orders : Array.isArray(json) ? json : [];

      if (!Array.isArray(orders)) {
        orders = [];
      }

      return orders.map((o: any) => ({
        id: o.id,
        customerId: o.user_id,
        customerName: o.user_name,
        items: (o.items || []).map((item: any) => ({
          productId: item.item_id,
          quantity: item.quantity,
        })),
        status: o.status?.toLowerCase() as 'pending' | 'processing' | 'completed' | undefined,
        totalPrice: o.order_price ? parseFloat(o.order_price) : undefined,
        createdAt: o.created_at,
        downloadUrl: o.download_url,
      }));
    } catch (error) {
      console.error('Error fetching orders:', error);
      throw error;
    }
  },

  async deleteOrder(orderId: string): Promise<void> {
    try {
      const response = await fetch(`${ORDERS_API_URL}/${orderId}`, {
        method: 'DELETE',
      });
      if (!response.ok) {
        throw new Error('Failed to delete order');
      }
    } catch (error) {
      console.error('Error deleting order:', error);
      throw error;
    }
  },

  // Create a new order
  async createOrder(order: Order): Promise<Order> {
    try {
      // Convert camelCase to snake_case for API
      const payload = {
        user_name: order.customerName,
        items: order.items.map((item) => ({
          item_id: item.productId,
          quantity: item.quantity,
        })),
      };

      const response = await fetch(`${ORDERS_API_URL}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error('Failed to create order');
      }
      const json = await response.json();

      // Map response from API to Order interface
      let createdOrder = Array.isArray(json.data) ? json.data[0] : json;

      if (!createdOrder) {
        throw new Error('Invalid response: no order data returned');
      }

      return {
        id: createdOrder.id,
        customerId: createdOrder.customer_id,
        customerName: createdOrder.user_name,
        items: (createdOrder.items || []).map((item: any) => ({
          productId: item.product_id,
          quantity: item.quantity,
        })),
        status: createdOrder.status?.toLowerCase() as 'pending' | 'processing' | 'completed' | undefined,
        totalPrice: createdOrder.order_price ? parseFloat(createdOrder.order_price) : undefined,
        createdAt: createdOrder.created_at,
        downloadUrl: createdOrder.download_url,
      };
    } catch (error) {
      console.error('Error creating order:', error);
      throw error;
    }
  },
};
