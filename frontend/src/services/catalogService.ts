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
  // Get presigned upload URL for blob storage
  // Backend returns both upload URL (for PUT) and download URL (for image display)
  // Request: POST /upload-url with { file_name, file_type }
  // Response: { upload_url: "...", download_url: "..." }
  async getPresignedUploadUrl(
    fileName: string,
    fileType: string
  ): Promise<{ uploadUrl: string; downloadUrl: string }> {
    try {
      const response = await fetch(`${CATALOG_API_URL}/upload-url`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          file_name: fileName,
          file_type: fileType,
        }),
      });

      if (!response.ok) {
        throw new Error('Failed to get presigned upload URL');
      }

      const json = await response.json();
      return {
        uploadUrl: json.upload_url || json.uploadUrl || '',
        downloadUrl: json.download_url || json.downloadUrl || '',
      };
    } catch (error) {
      console.error('Error getting presigned upload URL:', error);
      throw error;
    }
  },

  // Fetch all products from the catalog
  async getAllProducts(): Promise<Product[]> {
    try {
      const response = await fetch(`${CATALOG_API_URL}`);
      if (!response.ok) {
        throw new Error('Failed to fetch products');
      }
      const json = await response.json();

      // API returns { data: [...], count: N, timestamp: ... }
      // Extract data array and map field names from snake_case to camelCase
      let products = Array.isArray(json.data) ? json.data : Array.isArray(json) ? json : [];

      if (!Array.isArray(products)) {
        products = [];
      }

      return products.map((p: any) => ({
        id: p.id,
        code: p.code,
        name: p.name,
        image: p.image_url || '',
        price: parseFloat(p.price),
        stockQuantity: p.stock_quantity || 0,
      }));
    } catch (error) {
      console.error('Error fetching products:', error);
      throw error;
    }
  },

  // Create a new product
  async createProduct(product: Product): Promise<Product> {
    try {
      // Convert camelCase to snake_case for API
      const payload = {
        code: product.code,
        name: product.name,
        image_url: product.image,
        price: product.price,
        stock_quantity: product.stockQuantity,
      };

      const response = await fetch(`${CATALOG_API_URL}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });
      if (!response.ok) {
        throw new Error('Failed to create product');
      }
      const json = await response.json();

      // Map response from API to Product interface
      let createdProduct = Array.isArray(json.data) ? json.data[0] : json;

      if (!createdProduct) {
        throw new Error('Invalid response: no product data returned');
      }

      return {
        id: createdProduct.id,
        code: createdProduct.code,
        name: createdProduct.name,
        image: createdProduct.image || '',
        price: parseFloat(createdProduct.price),
        stockQuantity: createdProduct.stock_quantity || 0,
      };
    } catch (error) {
      console.error('Error creating product:', error);
      throw error;
    }
  },
};
