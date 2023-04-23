import React from "react";
import { Link } from "react-router-dom";
import "./Home.css";

const Home = () => {
  return (
    <div className="home-container">
      <h1 className="home-title">Package Manager</h1>
      <ul className="home-nav">
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/Search">
            Search
          </Link>
        </li>
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/upload">
            Upload
          </Link>
        </li>
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/update">
            Update
          </Link>
        </li>
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/delete">
            Delete
          </Link>
        </li>
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/create">
            Create
          </Link>
        </li>
        <li className="home-nav-item">
          <Link className="home-nav-link" to="/rate">
            Rate
          </Link>
        </li>
      </ul>
    </div>
  );
};

export default Home;
