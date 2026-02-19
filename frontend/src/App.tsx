import './App.css'
import PostList from './components/PostList'
import Header from './components/Header'
import { BrowserRouter, Routes, Route } from "react-router-dom"
import ProductsList from './features/products/components/ProductsList'
import { Toaster } from "@/components/ui/sonner"
import { ThemeProvider } from './components/theme-provider'
import { AuthProvider } from './context/AuthContext'
import CartDrawer from './features/cart/CartDrawer'
import AuthForm from './features/auth/components/AuthForm'

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
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
          <Route path="/login" element={
            <div className="container mx-auto py-16">
              <AuthForm />
            </div>} />
        </Routes>
        <Toaster />
      </BrowserRouter>
      </AuthProvider>
    </ThemeProvider>
  )
}

export default App
