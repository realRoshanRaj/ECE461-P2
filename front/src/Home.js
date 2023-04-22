import React from "react";
import {Link} from "react-router-dom";
// importing Link from react-router-dom to navigate to 
// different end points.
  
const Home = () => {
  return (
    <div>
      <h1>Package Manager</h1>
      <br />
      <ul>
        <li>
          {/* Endpoint to route to Home component */}
          <Link to="/Search">Search</Link>
        </li>
        <li>
          {/* Endpoint to route to About component */}
          <Link to="/upload">Upload</Link>
        </li>
        <li>
          {/* Endpoint to route to Contact Us component */}
          <Link to="/update">Update</Link>
        </li>
        <li>
        <Link to="/delete">Delete</Link>
        </li>
        <li>
        <Link to="/create">create</Link>
        </li>
        <li>
        <Link to="/rate">rate</Link>
        </li>
      </ul>
    </div>
  );
};
  
export default Home;