import './App.css'
import PostList from './components/PostList'
import Header from './components/Header'
import { BrowserRouter, Routes, Route } from "react-router-dom";
import ProductsList from './features/products/components/ProductsList';

function App() {
  return (
    <BrowserRouter>
      <Header />
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
    </BrowserRouter>
  )
}

export default App
