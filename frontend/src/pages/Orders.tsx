import { useState, useEffect } from 'react';
import { Button } from '../components/ui/button';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../components/ui/table';
import { CreateOrderModal } from '../components/modals/CreateOrderModal';
import { ordersService } from '../services/ordersService';
import type { Order } from '../services/ordersService';
import { Download, ShoppingCart } from 'lucide-react';

export const Orders = () => {
  const [orders, setOrders] = useState<Order[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch orders on component mount
  useEffect(() => {
    fetchOrders();
  }, []);

  const fetchOrders = async () => {
    try {
      setIsLoading(true);
      setError(null);
      console.log('Fetching orders from:', import.meta.env.VITE_ORDERS_API_URL);
      const data = await ordersService.getAllOrders();
      console.log('Orders fetched:', data);
      setOrders(Array.isArray(data) ? data : []);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to fetch orders';
      console.error('Error fetching orders:', err);
      setError(errorMessage);
      setOrders([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateOrder = async (order: Order) => {
    try {
      setIsSubmitting(true);
      const newOrder = await ordersService.createOrder(order);
      setOrders((prev) => [...prev, newOrder]);
      setIsModalOpen(false);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to create order';
      setError(errorMessage);
      console.error('Error creating order:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  // Get status badge color
  const getStatusBadgeColor = (status?: string) => {
    switch (status) {
      case 'pending':
        return 'bg-yellow-100 text-yellow-800';
      case 'processing':
        return 'bg-blue-100 text-blue-800';
      case 'completed':
        return 'bg-green-100 text-green-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const handleDownloadPDF = (downloadUrl?: string) => {
    if (!downloadUrl) {
      alert('PDF not available yet. Please wait for the order to be processed.');
      return;
    }
    // Open the signed Azure Blob Storage URL in a new tab
    window.open(downloadUrl, '_blank');
  };

  return (
    <div className="flex flex-col gap-6 p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Orders</h1>
          <p className="text-gray-500 mt-1">
            View and manage all orders
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>
          Create Order
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-800">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchOrders}
            className="mt-2"
          >
            Try Again
          </Button>
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex justify-center items-center h-64">
          <p className="text-gray-500">Loading orders...</p>
        </div>
      )}

      {/* Orders Table */}
      {!isLoading && orders.length > 0 && (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader className="bg-gray-50">
              <TableRow>
                <TableHead className="font-semibold">Order ID</TableHead>
                <TableHead className="font-semibold">Customer Name</TableHead>
                <TableHead className="font-semibold text-right">
                  Total Price
                </TableHead>
                <TableHead className="font-semibold">Status</TableHead>
                <TableHead className="font-semibold text-center">
                  Action
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {orders.map((order) => (
                <TableRow key={order.id} className="hover:bg-gray-50">
                  <TableCell className="font-medium">{order.id}</TableCell>
                  <TableCell>{order.customerName}</TableCell>
                  <TableCell className="text-right">
                    {order.totalPrice
                      ? `$${order.totalPrice.toFixed(2)}`
                      : 'N/A'}
                  </TableCell>
                  <TableCell>
                    <span
                      className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getStatusBadgeColor(order.status)}`}
                    >
                      {order.status || 'Unknown'}
                    </span>
                  </TableCell>
                  <TableCell className="text-center">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => handleDownloadPDF(order.downloadUrl)}
                      disabled={!order.downloadUrl}
                      title={
                        !order.downloadUrl
                          ? 'PDF available only when order is completed'
                          : 'Download invoice PDF'
                      }
                    >
                      <Download className="w-4 h-4" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && orders.length === 0 && !error && (
        <div className="empty-state">
          <div className="empty-state-icon">
            <ShoppingCart className="w-12 h-12 mx-auto opacity-50" />
          </div>
          <h2 className="empty-state-title">No Orders</h2>
          <p className="empty-state-description">
            No orders created yet. Start by creating your first order.
          </p>
          <Button onClick={() => setIsModalOpen(true)}>
            Create First Order
          </Button>
        </div>
      )}

      {/* Create Order Modal */}
      <CreateOrderModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleCreateOrder}
        isLoading={isSubmitting}
      />
    </div>
  );
};
