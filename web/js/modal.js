// Get the modal
var modal = document.getElementById("myModal");

// Get the buttons to show the modal
var exitButton = document.getElementById("exitButton");
var replayButton = document.getElementById("replayButton");

// When the game is over, display the modal
// For example, you can use a function like showGameOverModal() to trigger this
function showGameOverModal(winner, playerSeat) {
  modal.style.display = "flex";
  document.getElementById("winner").innerHTML = winner.seat == playerSeat
    ? "You win"
    : "You lose";
}

function hideGameOverModal() {
  modal.style.display = "none";
}
// Add event listeners to the buttons
exitButton.addEventListener("click", function () {
  modal.style.display = "none";
  window.location = "/";
  // Add code to exit the game or perform other actions
});

replayButton.addEventListener("click", function () {
  modal.style.display = "none";
  conn.send(JSON.stringify({
    up: false,
    down: false,
    start: false,
    replay: true,
  }));
});
