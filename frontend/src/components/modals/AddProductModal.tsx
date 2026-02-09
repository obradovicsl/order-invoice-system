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
import { catalogService } from '../../services/catalogService';
import { Upload } from 'lucide-react';

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
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [dragActive, setDragActive] = useState(false);

  // Validate form data
  const validateForm = (): boolean => {
    const newErrors: Record<string, string> = {};

    if (!formData.code.trim()) {
      newErrors.code = 'Code is required';
    }
    if (!formData.name.trim()) {
      newErrors.name = 'Name is required';
    }
    if (!selectedFile) {
      newErrors.image = 'Image is required';
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

  const ALLOWED_FORMATS = ['image/png', 'image/jpeg', 'image/webp'];

  const handleFileSelect = (file: File | null) => {
    if (!file) {
      setSelectedFile(null);
      return;
    }

    if (!ALLOWED_FORMATS.includes(file.type)) {
      setErrors((prev) => ({
        ...prev,
        image: 'Only PNG, JPEG, and WebP images are allowed',
      }));
      return;
    }

    setSelectedFile(file);
    // Clear image error when file is selected
    if (errors.image) {
      setErrors((prev) => {
        const newErrors = { ...prev };
        delete newErrors.image;
        return newErrors;
      });
    }
  };

  const handleDrag = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  };

  const handleDrop = (e: React.DragEvent<HTMLDivElement>) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);

    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (files && files.length > 0) {
      handleFileSelect(files[0]);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) {
      return;
    }

    try {
      if (!selectedFile) {
        throw new Error('No file selected');
      }

      // Step 1: Get presigned upload and download URLs from backend
      const { uploadUrl, downloadUrl } = await catalogService.getPresignedUploadUrl(
        selectedFile.name,
        selectedFile.type
      );

      // Step 2: Upload image to blob storage using presigned upload URL
      await fetch(uploadUrl, {
        method: 'PUT',
        headers: {
          'Content-Type': selectedFile.type,
          'x-ms-blob-type': 'BlockBlob',
        },
        body: selectedFile,
      });

      // Step 3: Create product with download URL as image
      const productData: Product = {
        ...formData,
        image: downloadUrl,
      };

      await onSubmit(productData);

      // Reset form after successful submission
      setFormData({
        code: '',
        name: '',
        image: '',
        price: 0,
        stockQuantity: 0,
      });
      setSelectedFile(null);
      onClose();
    } catch (error) {
      const errorMessage =
        error instanceof Error ? error.message : 'Error uploading image';
      setErrors((prev) => ({
        ...prev,
        image: errorMessage,
      }));
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

          {/* Image Upload Field */}
          <div className="space-y-2">
            <Label>Product Image</Label>
            <div
              onDragEnter={handleDrag}
              onDragLeave={handleDrag}
              onDragOver={handleDrag}
              onDrop={handleDrop}
              className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
                dragActive
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-300 hover:border-gray-400'
              } ${isLoading ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
            >
              <input
                type="file"
                id="image-upload"
                className="hidden"
                accept="image/png,image/jpeg,image/webp"
                onChange={handleFileInputChange}
                disabled={isLoading}
              />
              <label
                htmlFor="image-upload"
                className="flex flex-col items-center gap-2 cursor-pointer"
              >
                <Upload className="w-8 h-8 text-gray-400" />
                <div>
                  <p className="text-sm font-medium text-gray-700">
                    {selectedFile ? selectedFile.name : 'Drag and drop your image here'}
                  </p>
                  <p className="text-xs text-gray-500 mt-1">
                    or click to select (PNG, JPEG, WebP)
                  </p>
                </div>
              </label>
            </div>
            {errors.image && (
              <p className="text-sm text-red-500">{errors.image}</p>
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
