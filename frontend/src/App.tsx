import './App.css'
import PostList from './components/PostList'
import Header from './components/Header'
import { BrowserRouter, Routes, Route } from "react-router-dom"
import ProductsList from './features/products/components/ProductsList'
import { Toaster } from "@/components/ui/sonner"
import { ThemeProvider } from './components/theme-provider'
import CartDrawer from './features/cart/CartDrawer'

function App() {
  return (
    <ThemeProvider>
      <BrowserRouter>
        <Header />
        <CartDrawer />
        <Routes>
          <Route path="/" element={
            <div className="container mx-auto py-16">
              asdasdasd
              {/* <PostList /> */}
            </div>} />

          <Route path="/products" element={
            <div className="container mx-auto py-16">
              <ProductsList />
            </div>} />
          <Route path="/posts" element={
            <div className="container mx-auto py-16">
              <PostList />
            </div>} />
        </Routes>
        <Toaster />
      </BrowserRouter>
    </ThemeProvider>
  )
}

export default App
