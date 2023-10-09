# LooLock Telegram Bot 🤖

Welcome to `loolock-tg-bot`! This repository contains the source code for a Telegram bot written in Go.

## Description

LooLock is a Telegram bot designed to [short description of the bot's main function]. Built with Go, it offers fast and efficient performance for all your bot needs.

## Getting Started

### Prerequisites

- Go (v1.xx.x or newer)
- [Telegram API Token](https://core.telegram.org/bots#creating-a-new-bot) from @BotFather

### Installation

1. Clone the repository:

```
git clone https://github.com/yourusername/loolock-tg-bot.git
cd loolock-tg-bot
```

2. Create `.env` file with the right params

```
BOT_TOKEN=
DATABASE_URL=
```

3. Create the database

```
migrate -database DATABASE_URL -path db/migrations up
```

4. Run the bot

```
go run main.go
```

## Usage

Create your bot using @BotFather first and then connect it to this.

## Contributing

Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.


## Future TODOs: 

[] add input from admin to determine  
  [] leaderboard mode  
  [] minimum lock amount X  
[] update users  
  [] check new leaderboard  
  [] update db  
  [] kick users who don't belong anymore  

hardening:  
[] start off when joining new group iterating through each member asking for their sigs  
[] leave if no longer admin  
[x] thank for making it admin  


other:  
[] only receive messages from lockers  