# Multiplayer Pong

A classic Pong game implemented in [Golang](https://golang.org/) with the added excitement of real-time multiplayer gameplay powered by WebSockets. Challenge your friends to a match of Pong and see who can score the most points in this fast-paced game!

## Installation

1. Clone this repository to your local machine.

2. Open a terminal and navigate to the project directory.

3. Build the Docker image:

    ```sh
    docker build -t pong .
    ```

4. Run the Docker container:

    ```sh
    docker run -p 8080:8080 pong
    ```

5. The Go application will be available at http://localhost:8080.

## Directory Structure

- `cmd/`: Contains the application's main code.
- `internal/`: Contains internal packages and modules.
- `web/`: Contains static files for the web application, such as HTML, CSS, and JavaScript.

## Real-Time Multiplayer

This Pong game leverages WebSocket technology to enable real-time multiplayer gameplay. Players can connect to the game from different devices and challenge each other in a fast-paced match. The game operates with a 60 tick rate, ensuring that game updates and player inputs are processed at a high frequency. This results in a smooth and responsive gaming experience, making every move count.
