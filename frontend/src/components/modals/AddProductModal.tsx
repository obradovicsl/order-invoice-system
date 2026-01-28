import { useState } from 'react';
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
import type { Product } from '../../services/catalogService';

interface AddProductModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (product: Product) => Promise<void>;
  isLoading?: boolean;
}

export const AddProductModal = ({
  isOpen,
  onClose,
  onSubmit,
  isLoading = false,
}: AddProductModalProps) => {
  const [formData, setFormData] = useState<Product>({
    code: '',
    name: '',
    image: '',
    price: 0,
    stockQuantity: 0,
  });

  const [errors, setErrors] = useState<Record<string, string>>({});

  // Validate form data
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.code.trim()) {
      newErrors.code = 'Code is required';
    }
    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!formData.image.trim()) {
      newErrors.image = 'Image URL is required';
    }
    if (formData.price <= 0) {
      newErrors.price = 'Price must be greater than 0';
    }
    if (formData.stockQuantity <= 0) {
      newErrors.stockQuantity = 'Stock quantity must be greater than 0';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]:
        type === 'number' ? (value ? parseFloat(value) : 0) : value,
    }));
    // Clear error for this field when user starts typing
    if (errors[name]) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors[name];
        return newErrors;
      });
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) {
      return;
    }

    try {
      await onSubmit(formData);
      // Reset form after successful submission
      setFormData({
        code: '',
        name: '',
        image: '',
        price: 0,
        stockQuantity: 0,
      });
      onClose();
    } catch (error) {
      console.error('Error submitting form:', error);
    }
  };

  return (
    <Dialog open={isOpen} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[425px]">
        <DialogHeader>
          <DialogTitle>Add New Product</DialogTitle>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* Code Field */}
          <div className="space-y-2">
            <Label htmlFor="code">Code</Label>
            <Input
              id="code"
              name="code"
              placeholder="Enter product code"
              value={formData.code}
              onChange={handleInputChange}
              disabled={isLoading}
            />
            {errors.code && (
              <p className="text-sm text-red-500">{errors.code}</p>
            )}
          </div>

          {/* Name Field */}
          <div className="space-y-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              name="name"
              placeholder="Enter product name"
              value={formData.name}
              onChange={handleInputChange}
              disabled={isLoading}
            />
            {errors.name && (
              <p className="text-sm text-red-500">{errors.name}</p>
            )}
          </div>

          {/* Image Field */}
          <div className="space-y-2">
            <Label htmlFor="image">Image URL</Label>
            <Input
              id="image"
              name="image"
              placeholder="Enter image URL"
              value={formData.image}
              onChange={handleInputChange}
              disabled={isLoading}
            />
            {errors.image && (
              <p className="text-sm text-red-500">{errors.image}</p>
            )}
          </div>

          {/* Price Field */}
          <div className="space-y-2">
            <Label htmlFor="price">Price</Label>
            <Input
              id="price"
              name="price"
              type="number"
              placeholder="Enter product price"
              step="0.01"
              min="0.01"
              value={formData.price || ''}
              onChange={handleInputChange}
              disabled={isLoading}
            />
            {errors.price && (
              <p className="text-sm text-red-500">{errors.price}</p>
            )}
          </div>

          {/* Stock Quantity Field */}
          <div className="space-y-2">
            <Label htmlFor="stockQuantity">Stock Quantity</Label>
            <Input
              id="stockQuantity"
              name="stockQuantity"
              type="number"
              placeholder="Enter stock quantity"
              min="1"
              value={formData.stockQuantity || ''}
              onChange={handleInputChange}
              disabled={isLoading}
            />
            {errors.stockQuantity && (
              <p className="text-sm text-red-500">
                {errors.stockQuantity}
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
              {isLoading ? 'Adding...' : 'Add Product'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
};
