import React from "react";
//import axios from 'axios'

const Search = () => {
        const payload = {
          Version: "~4.18.2",
          Name: "express",
        };
    
        fetch("http://localhost:8080/packages", {
          method: "POST",
          mode: "no-cors",
          body: JSON.stringify([payload]),
        })
          .then((response) => response.json())
          .then((data) => console.log(data)
          .catch((error) => console.error(error)));
      

    return (
        
        <div>
            <input type="text" placeholder="Version" />
            <button>Search</button>
            <br />
            <input type="text" placeholder="Name" />
        </div>
    );
    };

export default Search;