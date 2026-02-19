import { Link } from 'react-router-dom'
import reactLogo from '../assets/react.svg'
import { ModeToggle } from './theme-toggle'
import { Button } from '@/components/ui/button'
import { ShoppingCart } from 'lucide-react'
import { useCartStore } from '@/features/cart/store'

function Header() {
  const { openCart, items } = useCartStore()
  const itemCount = items.reduce((n, i) => n + i.quantity, 0)

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
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            onClick={openCart}
            className="relative text-gray-300 hover:text-white"
            aria-label="Open cart"
          >
            <ShoppingCart className="size-5" />
            {itemCount > 0 && (
              <span className="absolute right-0 top-0 flex size-4 items-center justify-center rounded-full bg-primary text-[10px] font-medium text-primary-foreground">
                {itemCount > 99 ? "99+" : itemCount}
              </span>
            )}
          </Button>
          <ModeToggle />
        </div>
      </div>
    </header>
  )
}

export default Header
