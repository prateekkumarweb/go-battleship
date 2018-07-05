var socketProtocol = "ws:"

if (document.location.protocol == "https:") {
  socketProtocol = "wss:"
}

var socket = new WebSocket(socketProtocol+"//"+document.location.host+"/ws");

$(function() {
  /**
   * Successfully connected to server event
   */
  socket.onopen = function() {
    console.log('Connected to server.');
    $('#disconnected').hide();
    $('#waiting-room').show();
  }

  /**
   * Disconnected from server event
   */
  socket.onclose = function() {
    console.log('Disconnected from server.');
    $('#waiting-room').hide();
    $('#game').hide();
    $('#disconnected').show();
  }

  socket.onmessage = function(e) {
    var data = JSON.parse(e.data)
    var evt = data.event
    var message = data.message

    if (evt == "join") {
      var gameId = message;
      Game.initGame();
      $('#messages').empty();
      $('#disconnected').hide();
      $('#waiting-room').hide();
      $('#game').show();
      $('#game-number').html(gameId);
      return;
    }

    if (evt == "update") {
      var gameState = JSON.parse(message);
      Game.setTurn(gameState.turn);
      Game.updateGrid(gameState.gridIndex, gameState.grid);
      return;
    }

    if (evt == "chat") {
      var msg = JSON.parse(message);
      $('#messages').append('<li><strong>' + msg.name + ':</strong> ' + msg.message + '</li>');
      $('#messages-list').scrollTop($('#messages-list')[0].scrollHeight);
      return;
    }

    if (evt == "notification") {
      var msg = JSON.parse(message);
      $('#messages').append('<li>' + msg.message + '</li>');
      $('#messages-list').scrollTop($('#messages-list')[0].scrollHeight);
      return;
    }

    if (evt == "gameover") {
      var isWinner = JSON.parse(message);
      Game.setGameOver(isWinner.isWinner);
      return;
    }

    if (evt == "leave") {
      $('#game').hide();
      $('#waiting-room').show();
      return;
    }
  }
  //
  // /**
  //  * User has joined a game
  //  */
  //
  // socket.addEventListener('join', function(gameId) {
  //   Game.initGame();
  //   $('#messages').empty();
  //   $('#disconnected').hide();
  //   $('#waiting-room').hide();
  //   $('#game').show();
  //   $('#game-number').html(gameId);
  // })
  //
  // /**
  //  * Update player's game state
  //  */
  // socket.addEventListener('update', function(gameState) {
  //   Game.setTurn(gameState.turn);
  //   Game.updateGrid(gameState.gridIndex, gameState.grid);
  // });
  //
  // /**
  //  * Game chat message
  //  */
  // socket.addEventListener('chat', function(msg) {
  //   $('#messages').append('<li><strong>' + msg.name + ':</strong> ' + msg.message + '</li>');
  //   $('#messages-list').scrollTop($('#messages-list')[0].scrollHeight);
  // });
  //
  // /**
  //  * Game notification
  //  */
  // socket.addEventListener('notification', function(msg) {
  //   $('#messages').append('<li>' + msg.message + '</li>');
  //   $('#messages-list').scrollTop($('#messages-list')[0].scrollHeight);
  // });
  //
  // /**
  //  * Change game status to game over
  //  */
  // socket.addEventListener('gameover', function(isWinner) {
  //   Game.setGameOver(isWinner);
  // });
  //
  // /**
  //  * Leave game and join waiting room
  //  */
  // socket.addEventListener('leave', function() {
  //   $('#game').hide();
  //   $('#waiting-room').show();
  // });

  /**
   * Send chat message to server
   */
  $('#message-form').submit(function() {
    socket.send(JSON.stringify({event: 'chat', message: $('#message').val()}))
    // socket.emit('chat', $('#message').val());
    $('#message').val('');
    return false;
  });
});

/**
 * Send leave game request
 * @param {type} e Event
 */
function sendLeaveRequest(e) {
  e.preventDefault();
  socket.send(JSON.stringify({event: 'leave', message: ''}))
  // socket.emit('leave');
}

/**
 * Send shot coordinates to server
 * @param {type} square
 */
function sendShot(square) {
  socket.send(JSON.stringify({event: 'shot', message: JSON.stringify(square)}))
  // socket.emit('shot', square);
}
