# --------------------
# Stage 1: Build the UI
# --------------------
    FROM node:22-alpine AS build-ui

    # Set working directory and copy the project files.
    WORKDIR /app
    # Copy all files; you may want to add a .dockerignore to skip ui/node_modules etc.
    COPY . .
    
    # Change to the UI folder, install dependencies, and build.
    WORKDIR /app/ui
    RUN npm install && npm run build
    
    # --------------------
    # Stage 2: Build the Go binary for multiple platforms
    # --------------------
    FROM golang:1.23 AS build-go
    
    # Set working directory and copy the project.
    WORKDIR /app
    COPY . .
    
    # Copy the built UI from the previous stage into the proper location.
    COPY --from=build-ui /app/ui/dist ./ui/dist
    
    # Create an output directory for the binaries.
    RUN mkdir -p bin

    RUN CGO_ENABLED=0 go build

    # Build for Linux (amd64)
    RUN GOOS=linux GOARCH=amd64 go build -o bin/myapp-linux-amd64 .


    RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o bin/myapp-linux-arm64 .
    
    
    # Build for macOS (darwin/arm64)
    RUN GOOS=darwin GOARCH=arm64 go build -o bin/myapp-darwin-arm64 .
    


    # --------------------
    # Optionally, you can define a final stage to package or export the results
    # --------------------
    FROM alpine:latest AS final
    WORKDIR /app
    # Copy the compiled binaries from the build stage.
    COPY --from=build-go /app/bin ./bin
    COPY --from=build-go /app/motisConfigServer ./motisConfigServer
    
    # This stage just holds the binaries; you can add further instructions to package or run tests.
    CMD ["sh"]
    