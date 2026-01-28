import { Link, useLocation } from 'react-router-dom';
import { Button } from './ui/button';

export const Navbar = () => {
  const location = useLocation();

  const isActive = (path: string) => location.pathname === path;

  return (
    <nav className="bg-white shadow-md">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo / Brand */}
          <div className="flex-shrink-0">
            <Link to="/" className="font-bold text-xl text-gray-900">
              Order System
            </Link>
          </div>

          {/* Navigation Links */}
          <div className="flex space-x-4">
            <Link to="/catalog">
              <Button
                variant={isActive('/catalog') ? 'default' : 'ghost'}
                className="font-medium"
              >
                Catalog
              </Button>
            </Link>

            <Link to="/orders">
              <Button
                variant={isActive('/orders') ? 'default' : 'ghost'}
                className="font-medium"
              >
                Orders
              </Button>
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
};
