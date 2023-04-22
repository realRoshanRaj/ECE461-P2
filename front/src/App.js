import {BrowserRouter as Router, Route, Routes} from "react-router-dom";
// import Home component
import Home from "./Home";
import Search from "./components/Search";
// import About component
  
function App() {
  return (
    <Router>
      <Routes>
        <Route exact path="/" element={<Home />} />
        <Route exact path="/search" element={<Search />} />
      </Routes>
    </Router>
  )
}
  
export default App;