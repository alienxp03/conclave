import { Link, useLocation } from 'react-router-dom';

export function Navigation() {
  const location = useLocation();

  return (
    <nav className="bg-gray-800 shadow-lg border-b border-gray-700 sticky top-0 z-50 backdrop-blur-sm bg-opacity-95">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between h-16">
          <div className="flex items-center">
            <Link to="/" className="flex items-center space-x-2 group">
              <span className="text-2xl transform transition-transform group-hover:scale-110">
                🎭
              </span>
              <span className="font-bold text-xl text-white">dbate</span>
            </Link>
          </div>
          <div className="flex items-center space-x-4">
            <Link
              to="/history"
              className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                location.pathname === '/history'
                  ? 'bg-gray-700 text-white'
                  : 'text-gray-300 hover:text-white hover:bg-gray-700'
              }`}
            >
              History
            </Link>
          </div>
        </div>
      </div>
    </nav>
  );
}
