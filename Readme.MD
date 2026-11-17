🥪 PB&J – League of Legends Pick/Ban Helper

PB&J is a League of Legends draft assistant designed to help players make better decisions during champion select.

It tracks champion picks and bans throughout a draft and evaluates enemy champion selections against your preferred champions, helping you identify favourable matchups.

✨ Features
Champion selection tracking – Keep track of picks and bans for both Blue and Red side.
Player profiles – Register players with their preferred champions and assigned role.
Champion matchups – Retrieve win-rate information for your preferred champions against enemy picks.
Role-aware statistics – Account for the lane in which an opposing champion is being played.
Automatic champion data – Retrieve champion information without manually maintaining a champion list.

🛠️ Tech Stack
Go – Backend implementation.
Huma – API framework and OpenAPI support.
OP.GG – Champion matchup statistics.
🚀 Getting Started
Prerequisites
Go installed on your machine.
An internet connection for retrieving champion and matchup data.
Installation

Clone the repository:

```bash
git clone https://github.com/dessertivore/pb-and-j.git
cd pb-and-j
```

Download dependencies:

```bash
go mod download
```

Running the application

Start the API server:

```bash
go run .
```

The API listens on:

```text
http://127.0.0.1:8888
```
📡 API Reference

PB&J exposes a REST API for managing champions, players and draft sessions.

Method	Endpoint	Description
GET	/champions	Retrieve available champions
GET	/user	Retrieve registered users
POST	/user	Register a user
POST	/game	Create a draft session
GET	/game	Retrieve draft sessions
POST	/action	Submit a pick or ban

A typical draft session follows this process:

Retrieve the available champions.
Register players with their preferred champions and assigned roles.
Create a game, specifying the player and their side (Blue or Red).
Submit picks and bans as the draft progresses.
When an opposing champion is selected, retrieve matchup win-rate information for your preferred champions.
Example: Register a player

POST /user

```json
{
"displayName": "Player One",
"username": "player1",
"preferredChampions": [
"smolder",
"jhin",
"ezreal"
],
"assignedRole": "bottom"
}
```
Example: Create a game

POST /game

```text
?username=player1&side=blue
```

The response contains a game ID which can be used to identify the draft session.

Example: Submit a draft action

POST /action

```json
{
"pick_ban_id": "YOUR_GAME_ID",
"action": {
"side": "red",
"type": "pick",
"champion": "jinx",
"role": "bottom"
}
}
```

The API can return matchup win-rate information for the player's preferred champions.

🗺️ Draft Structure

The helper follows the standard League of Legends draft sequence:

Phase	Actions
First ban phase	3 bans per side
First pick phase	3 picks per side
Second ban phase	2 bans per side
Final pick phase	2 picks per side

Both sides are tracked independently, allowing the helper to understand the current state of the draft.

📊 Matchup Analysis

When an opposing champion is selected, PB&J retrieves matchup statistics for the player's preferred champions.

The intention is to help answer questions such as:

Which of my preferred champions has the best matchup?
Which champion should I consider picking next?
How favourable is a particular matchup in my assigned role?

Matchup statistics are sourced from external champion statistics websites and may change over time.

⚠️ Current Limitations

PB&J is an ongoing project.
No data is persisted - all is stored in memory and lost when server is restarted.
No frontend yet!