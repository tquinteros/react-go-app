import { Link } from 'react-router-dom'
import reactLogo from '../assets/react.svg'
import { ModeToggle } from './theme-toggle'
function Header() {
  return (
    <header className="border-b border-gray-700 bg-[#242424] sticky top-0 z-10">
      <div className="container mx-auto px-4 flex items-center justify-between h-14">
        <Link to="/" className="flex items-center gap-2">
          <img src={reactLogo} className="h-8 w-auto" alt="Logo" />
        </Link>
        <nav className="flex items-center gap-6">
          <Link
            to="/"
            className="text-sm font-medium text-gray-300 hover:text-white transition-colors"
          >
            Home
          </Link>
          <Link
            to="/posts"
            className="text-sm font-medium text-gray-300 hover:text-white transition-colors"
          >
            Posts
          </Link>
          <Link
            to="/products"
            className="text-sm font-medium text-gray-300 hover:text-white transition-colors"
          >
            Products
          </Link>
        </nav>
        <ModeToggle />
      </div>
    </header>
  )
}

export default Header
