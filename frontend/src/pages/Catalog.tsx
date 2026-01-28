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
import { AddProductModal } from '../components/modals/AddProductModal';
import { catalogService } from '../services/catalogService';
import type { Product } from '../services/catalogService';
import { Package } from 'lucide-react';

export const Catalog = () => {
  const [products, setProducts] = useState<Product[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  // Fetch products on component mount
  useEffect(() => {
    fetchProducts();
  }, []);

  const fetchProducts = async () => {
    try {
      setIsLoading(true);
      setError(null);
      console.log('Fetching products from:', import.meta.env.VITE_CATALOG_API_URL);
      const data = await catalogService.getAllProducts();
      console.log('Products fetched:', data);
      setProducts(Array.isArray(data) ? data : []);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to fetch products';
      console.error('Error fetching products:', err);
      setError(errorMessage);
      setProducts([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleAddProduct = async (product: Product) => {
    try {
      setIsSubmitting(true);
      const newProduct = await catalogService.createProduct(product);
      setProducts((prev) => [...prev, newProduct]);
      setIsModalOpen(false);
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : 'Failed to add product';
      setError(errorMessage);
      console.error('Error adding product:', err);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex flex-col gap-6 p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Product Catalog</h1>
          <p className="text-gray-500 mt-1">
            Manage and view all products
          </p>
        </div>
        <Button onClick={() => setIsModalOpen(true)}>
          Add Product
        </Button>
      </div>

      {/* Error Message */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-md p-4">
          <p className="text-red-800">{error}</p>
          <Button
            variant="outline"
            size="sm"
            onClick={fetchProducts}
            className="mt-2"
          >
            Try Again
          </Button>
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex justify-center items-center h-64">
          <p className="text-gray-500">Loading products...</p>
        </div>
      )}

      {/* Products Table */}
      {!isLoading && products.length > 0 && (
        <div className="border rounded-lg overflow-hidden">
          <Table>
            <TableHeader className="bg-gray-50">
              <TableRow>
                <TableHead className="font-semibold">Code</TableHead>
                <TableHead className="font-semibold">Name</TableHead>
                <TableHead className="font-semibold">Price</TableHead>
                <TableHead className="font-semibold text-right">
                  Stock Quantity
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {products.map((product) => (
                <TableRow key={product.code} className="hover:bg-gray-50">
                  <TableCell className="font-medium">{product.code}</TableCell>
                  <TableCell>{product.name}</TableCell>
                  <TableCell>${product.price.toFixed(2)}</TableCell>
                  <TableCell className="text-right">
                    {product.stockQuantity}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}

      {/* Empty State */}
      {!isLoading && products.length === 0 && !error && (
        <div className="empty-state">
          <div className="empty-state-icon">
            <Package className="w-12 h-12 mx-auto opacity-50" />
          </div>
          <h2 className="empty-state-title">No Products</h2>
          <p className="empty-state-description">
            No products available yet. Start by adding your first product.
          </p>
          <Button onClick={() => setIsModalOpen(true)}>
            Add First Product
          </Button>
        </div>
      )}

      {/* Add Product Modal */}
      <AddProductModal
        isOpen={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        onSubmit={handleAddProduct}
        isLoading={isSubmitting}
      />
    </div>
  );
};
