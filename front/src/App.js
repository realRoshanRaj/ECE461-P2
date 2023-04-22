import {BrowserRouter as Router, Route, Routes} from "react-router-dom";
// import Home component
import Home from "./Home";
import Search from "./components/Search";
import Create from "./components/Create";
import Delete from "./components/Delete";
import Update from "./components/Update";
import Upload from "./components/Upload";
import Rate from "./components/Rate";
// import About component
  
function App() {
  return (
    <Router>
      <Routes>
        <Route exact path="/" element={<Home />} />
        <Route exact path="/search" element={<Search />} />
        <Route exact path="/create" element={<Create />} />
        <Route exact path="/delete" element={<Delete />} />
        <Route exact path="/update" element={<Update />} />
        <Route exact path="/upload" element={<Upload />} />
        <Route exact path="/rate" element={<Rate />} />

      </Routes>
    </Router>
  )
}
  
export default App;