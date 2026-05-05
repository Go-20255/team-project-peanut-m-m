# Monopoly Go

## What is this project?

## Environment Setup 

### Requirements

Linux environment (either with a VM or WSL2 if on Windows) with the following packages:
- Go >= 1.25.7
- Node >= 24.7.0
- Docker
- [Just](https://github.com/casey/just) >= 1.47.1 (Makefile alternative)

### Setup

After installing the above packages, to setup your environment, run the following commands:
1. `just setup-environment`: creates a .internal.env with passwords; installs node modules in frontend
2. `just redeploy-ephemeral-postgres`: deploys an ephemeral postgres server in docker for data storage
3. Open two terminals
    1. First terminal run `just run-backend`: Starts the backend server
    2. Second terminal run `just run-frontend`: Starts the frontend server (http://localhost:3000)

