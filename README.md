# ECE461 Project Part 2

# Main Contributors
1. Roshan Raj
2. Ayush Praharaj
3. Seungkeun Lee
4. Rajeev Sashti

## Part 1 Made by:
- Alonso Cestti
- Bartosz Stoppel
- Nandini Krishna
- Nahush Walvekar

# 

Built a web-based package management system that has following functionalities.  
- upload, update, and download packages.  
- Rating metric (see the details below).  
- Package ingestion.  
- Fetch directory of package. 
- Fetch package history. 
- Searching. 
- System reset. 
- Popularity tracking. 

#

Packages are rated based on the following metrics  
  
RampUp time that rates how easy a project is to get up and running
Correctness which rates how many issues have been closed in the repo. 
BusFactor which measures the amount of risk in the project maintainence by not sharing objectives with a team. 
Responsive Maintainence which rates the project on how much maintainence the project receives. 
License compatability which rates the projects compatability with LGPLv2.1 license.
Code quality score which measures the fraction of project code contributed through pull requests with a code review.
Version score which masrues the fraction of dependencies fixed to a particular major and minor version.

Finally the Net score of a repository is calculated using the following formula Net Score = (license compatibility score)* (correctness score + 3* responsiveness score + Bus Factor + 2 * low-ramp up time + code_quality_score + version_score) / 10
