// Catalog service for managing product catalog operations

export interface Product {
  id?: string;
  code: string;
  name: string;
  image: string;
  price: number;
  stockQuantity: number;
}

const CATALOG_API_URL = import.meta.env.VITE_CATALOG_API_URL;

export const catalogService = {
  // Fetch all products from the catalog
  async getAllProducts(): Promise<Product[]> {
    try {
      const response = await fetch(`${CATALOG_API_URL}`);
      if (!response.ok) {
        throw new Error('Failed to fetch products');
      }
      return await response.json();
    } catch (error) {
      console.error('Error fetching products:', error);
      throw error;
    }
  },

  // Create a new product
  async createProduct(product: Product): Promise<Product> {
    try {
      const response = await fetch(`${CATALOG_API_URL}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(product),
      });
      if (!response.ok) {
        throw new Error('Failed to create product');
      }
      return await response.json();
    } catch (error) {
      console.error('Error creating product:', error);
      throw error;
    }
  },
};
