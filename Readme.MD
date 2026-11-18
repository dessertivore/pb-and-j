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

Run using Docker:

```bash
make run_app
```

Stop:
```bash
make stop
```

While the backend and frontend servers are running, the frontend can be accessed on

```text
http://localhost:5173/
```


A typical draft session follows this process:

Retrieve the available champions.
Register players with their preferred champions and assigned roles.
Create a game, specifying the player and their side (Blue or Red).
Submit picks and bans as the draft progresses.
When an opposing champion is selected, retrieve matchup win-rate information for your preferred champions.


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
