import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { Navbar } from './components/Navbar';
import { Catalog } from './pages/Catalog';
import { Orders } from './pages/Orders';
import './App.css';

function App() {
  return (
    <Router>
      <div className="flex flex-col min-h-screen bg-white">
        {/* Navigation Bar */}
        <Navbar />

        {/* Main Content */}
        <main className="flex-grow">
          <Routes>
            {/* Catalog page */}
            <Route path="/catalog" element={<Catalog />} />

            {/* Orders page */}
            <Route path="/orders" element={<Orders />} />

            {/* Default redirect to catalog */}
            <Route path="/" element={<Navigate to="/catalog" replace />} />

            {/* 404 fallback */}
            <Route
              path="*"
              element={
                <div className="flex justify-center items-center h-96">
                  <div className="text-center">
                    <h1 className="text-4xl font-bold text-gray-900 mb-2">
                      404
                    </h1>
                    <p className="text-gray-500">Page not found</p>
                  </div>
                </div>
              }
            />
          </Routes>
        </main>
      </div>
    </Router>
  );
}

export default App;
