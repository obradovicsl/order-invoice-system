import { useState, useEffect } from 'react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '../ui/dialog';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Label } from '../ui/label';
import { X } from 'lucide-react';
import type { Order, OrderItem } from '../../services/ordersService';
import { catalogService } from '../../services/catalogService';
import type { Product } from '../../services/catalogService';

interface CreateOrderModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (order: Order) => Promise<void>;
  isLoading?: boolean;
}

export const CreateOrderModal = ({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: CreateOrderModalProps) => {
  const [customerName, setCustomerName] = useState('');
  const [items, setItems] = useState<OrderItem[]>([]);
  const [availableProducts, setAvailableProducts] = useState<Product[]>([]);
  const [errors, setErrors] = useState<Record<string, string>>({});

  // Fetch available products
  useEffect(() => {
    if (isOpen) {
      fetchProducts();
    }
  }, [isOpen]);

  const fetchProducts = async () => {
    try {
      const products = await catalogService.getAllProducts();
      setAvailableProducts(products);
    } catch (error) {
      console.error('Error fetching products:', error);
    }
  };

  // Validate form data
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!customerName.trim()) {
      newErrors.customerName = 'Customer Name is required';
    }
    if (items.length === 0) {
      newErrors.items = 'At least one product must be added';
    }
    if (items.some((item) => item.quantity <= 0)) {
      newErrors.items = 'All products must have quantity greater than 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleAddItem = () => {
    // Add empty item row
    setItems((prev) => [
      ...prev,
      { productId: '', quantity: 1 },
    ]);
  };

  const handleRemoveItem = (index: number) => {
    setItems((prev) => prev.filter((_, i) => i !== index));
  };

  const handleItemChange = (
    index: number,
    field: 'productId' | 'quantity',
    value: string | number,
  ) => {
    setItems((prev) => {
      const newItems = [...prev];

      // For quantity validation, use the current item's productId
      if (field === 'quantity') {
        const quantity = parseInt(String(value));
        const currentItem = newItems[index];

        // Only validate if a product is selected
        if (currentItem.productId) {
          const selectedProduct = availableProducts.find(
            (p) => p.id === currentItem.productId || p.code === currentItem.productId
          );
          const maxQuantity = selectedProduct?.stockQuantity || 0;

          if (quantity > maxQuantity) {
            return prev;
          }
        }
      }

      newItems[index] = {
        ...newItems[index],
        [field]: field === 'quantity' ? parseInt(String(value)) : value,
      };
      return newItems;
    });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) {
      return;
    }

    const order: Order = {
      customerName,
      items,
    };

    try {
      await onSubmit(order);
      // Reset form
      setCustomerName('');
      setItems([]);
      onClose();
    } catch (error) {
      console.error('Error submitting form:', error);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Create New Order</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Customer Name Field */}
          <div className="space-y-2">
            <Label htmlFor="customerName">Customer Name</Label>
            <Input
              id="customerName"
              placeholder="Enter customer name"
              value={customerName}
              onChange={(e) => {
                setCustomerName(e.target.value);
                if (errors.customerName) {
                  setErrors((prev) => {
                    const newErrors = { ...prev };
                    delete newErrors.customerName;
                    return newErrors;
                  });
                }
              }}
              disabled={isLoading}
            />
            {errors.customerName && (
              <p className="text-sm text-red-500">{errors.customerName}</p>
            )}
          </div>

          {/* Products Section */}
          <div className="space-y-2 border-t pt-4">
            <div className="flex justify-between items-center">
              <Label>Products</Label>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddItem}
                disabled={isLoading}
              >
                Add Product
              </Button>
            </div>

            {errors.items && (
              <p className="text-sm text-red-500">{errors.items}</p>
            )}

            {/* Items List */}
            <div className="space-y-3">
              {items.map((item, index) => (
                <div
                  key={index}
                  className="flex gap-2 items-end border rounded p-3 bg-gray-50"
                >
                  {/* Product ID Select */}
                  <div className="flex-1 space-y-1">
                    <Label htmlFor={`product-${index}`} className="text-xs">
                      Product
                    </Label>
                    <select
                      id={`product-${index}`}
                      className="w-full px-3 py-2 border border-gray-300 rounded-md text-sm"
                      value={item.productId}
                      onChange={(e) =>
                        handleItemChange(index, 'productId', e.target.value)
                      }
                      disabled={isLoading}
                    >
                      <option value="">Select product...</option>
                      {availableProducts.map((product) => (
                        <option key={product.id} value={product.id || product.code}>
                          {product.code} - {product.name}
                        </option>
                      ))}
                    </select>
                  </div>

                  {/* Quantity Input */}
                  <div className="w-32 space-y-1">
                    <Label htmlFor={`quantity-${index}`} className="text-xs">
                      Quantity
                    </Label>
                    <Input
                      id={`quantity-${index}`}
                      type="number"
                      min="1"
                      max={
                        availableProducts.find(
                          (p) => p.id === item.productId || p.code === item.productId
                        )?.stockQuantity || 0
                      }
                      value={item.quantity}
                      onChange={(e) =>
                        handleItemChange(index, 'quantity', e.target.value)
                      }
                      disabled={isLoading}
                      className="text-sm"
                    />
                    {item.productId && (
                      <p className="text-xs text-gray-500">
                        Max: {availableProducts.find(
                          (p) => p.id === item.productId || p.code === item.productId
                        )?.stockQuantity || 0}
                      </p>
                    )}
                  </div>

                  {/* Remove Button */}
                  <button
                    type="button"
                    onClick={() => handleRemoveItem(index)}
                    disabled={isLoading}
                    className="text-gray-400 hover:text-red-500 transition-colors p-1"
                    title="Remove product"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              ))}

              {items.length === 0 && (
                <p className="text-sm text-gray-500 italic">
                  No products added yet
                </p>
              )}
            </div>

            {/* No products available message */}
            {availableProducts.length === 0 && (
              <p className="text-sm text-red-500 italic mt-2">
                No products available in catalog. Please add products first.
              </p>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Creating...' : 'Create Order'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
