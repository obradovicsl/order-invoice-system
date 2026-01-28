// Orders service for managing order operations

export interface OrderItem {
  productCode: string;
  quantity: number;
}

export interface Order {
  id?: string;
  customerId: string;
  customerName: string;
  items: OrderItem[];
  status?: 'pending' | 'processing' | 'completed';
  totalPrice?: number;
  createdAt?: string;
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
      return await response.json();
    } catch (error) {
      console.error('Error fetching orders:', error);
      throw error;
    }
  },

  // Create a new order
  async createOrder(order: Order): Promise<Order> {
    try {
      const response = await fetch(`${ORDERS_API_URL}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(order),
      });
      if (!response.ok) {
        throw new Error('Failed to create order');
      }
      return await response.json();
    } catch (error) {
      console.error('Error creating order:', error);
      throw error;
    }
  },
};
