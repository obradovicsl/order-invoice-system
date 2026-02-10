import { describe, it, expect, beforeEach, vi } from 'vitest';
import { ordersService, type Order } from './ordersService';

// Mock fetch
global.fetch = vi.fn();

describe('ordersService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('getAllOrders', () => {
    it('should fetch and transform orders successfully', async () => {
      const mockApiResponse = {
        orders: [
          {
            id: '123',
            user_id: 'user-1',
            user_name: 'John Doe',
            items: [
              {
                item_id: 'item-1',
                quantity: 2,
              },
            ],
            status: 'PENDING',
            order_price: '99.99',
            created_at: '2024-02-10T10:00:00Z',
          },
        ],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockApiResponse,
      });

      const orders = await ordersService.getAllOrders();

      expect(orders).toHaveLength(1);
      expect(orders[0].id).toBe('123');
      expect(orders[0].customerName).toBe('John Doe');
      expect(orders[0].items).toHaveLength(1);
      expect(orders[0].totalPrice).toBe(99.99);
    });

    it('should handle fetch error gracefully', async () => {
      (global.fetch as any).mockRejectedValueOnce(new Error('Network error'));

      await expect(ordersService.getAllOrders()).rejects.toThrow();
    });

    it('should handle non-ok response status', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
      });

      await expect(ordersService.getAllOrders()).rejects.toThrow('Failed to fetch orders');
    });

    it('should return empty array when response has no orders', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      const orders = await ordersService.getAllOrders();

      expect(orders).toEqual([]);
    });
  });

  describe('createOrder', () => {
    it('should create an order successfully', async () => {
      const orderInput: Order = {
        customerName: 'Jane Doe',
        items: [
          {
            productId: 'prod-1',
            quantity: 3,
          },
        ],
      };

      const mockApiResponse = {
        id: '456',
        user_name: 'Jane Doe',
        customer_id: 'user-2',
        items: [
          {
            product_id: 'prod-1',
            quantity: 3,
          },
        ],
        status: 'PENDING',
        order_price: '149.97',
        created_at: '2024-02-10T11:00:00Z',
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => mockApiResponse,
      });

      const createdOrder = await ordersService.createOrder(orderInput);

      expect(createdOrder.id).toBe('456');
      expect(createdOrder.customerName).toBe('Jane Doe');
      expect(createdOrder.items).toHaveLength(1);
      expect(createdOrder.totalPrice).toBe(149.97);
    });

    it('should handle create order error', async () => {
      const orderInput: Order = {
        customerName: 'Jane Doe',
        items: [
          {
            productId: 'prod-1',
            quantity: 1,
          },
        ],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
      });

      await expect(ordersService.createOrder(orderInput)).rejects.toThrow('Failed to create order');
    });

    it('should handle invalid response from create order', async () => {
      const orderInput: Order = {
        customerName: 'Jane Doe',
        items: [
          {
            productId: 'prod-1',
            quantity: 1,
          },
        ],
      };

      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await expect(ordersService.createOrder(orderInput)).rejects.toThrow('Invalid response: no order data returned');
    });
  });

  describe('deleteOrder', () => {
    it('should delete an order successfully', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: true,
      });

      await expect(ordersService.deleteOrder('123')).resolves.toBeUndefined();

      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/123'),
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });

    it('should handle delete order error', async () => {
      (global.fetch as any).mockResolvedValueOnce({
        ok: false,
      });

      await expect(ordersService.deleteOrder('123')).rejects.toThrow('Failed to delete order');
    });
  });
});
