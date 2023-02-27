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

Built a Command line tool that can take in a github or npm url to a repository and rates them on 5 metrics. 
RampUp time that rates how easy a project is to get up and running.
Correctness which rates how many issues have been closed in the repo.
BusFactor which measures the amount of risk in the project maintainence by not sharing objectives with a team.
Responsive Maintainence which rates the project on how much maintainence the project receives
License compatability which rates the projects compatability with LGPLv2.1 license

Finally the Net score of a repository is calculated using the following formula Net Score = (license compatibility score)* (correctness score + 3* responsiveness score + Bus Factor + 2 * low-ramp up time) / 7
