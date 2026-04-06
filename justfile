
# Setup environment for running the monopoly servers
setup-environment:
    #!/usr/bin/env bash

    if [ ! -f ".internal.env" ] ; then
        printf "Missing .internal.env in backend/env. Generating secure password and JWT secret...\n"

        postgres_pass=$(openssl rand -base64 24 | tr -dc 'A-Za-z0-9' | head -c 24)

        touch .internal.env
        set -o noclobber
        echo "POSTGRES_PASSWORD=$postgres_pass" >> .internal.env
        echo "POSTGRES_PORT=1357" >> .internal.env

        printf "\n.internal.env file created and populated!\n"
        printf "You can change these values manually by going into the .internal.env file 
    manually and editing the file"
    fi
    if [ ! -d "frontend/node_modules" ]; then
        printf "Missing node_modules for frontend; downloading dependencies...\n"
        cd frontend
        npm install
        cd ..
    fi
    printf "\nCompleted setup!\n"

# Run the backend server
[working-directory: 'backend']
run-backend:
    go run main.go

# Run the frontend server
[working-directory: 'frontend']
run-frontend:
    npm run dev

# Redeploy a local ephemeral postgres instance for dev
redeploy-ephemeral-postgres:
    #!/usr/bin/env bash
    source .internal.env
    docker stop monopoly-postgres-e
    docker rm monopoly-postgres-e
    echo "port: $POSTGRES_PORT, Pass: $POSTGRES_PASSWORD"
    docker run --name monopoly-postgres-e -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD -d -p $POSTGRES_PORT:5432 postgres
