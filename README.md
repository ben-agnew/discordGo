# Discord Rank Bot

A simple go bot for taking user slash commands and returing the users rank data in either Rocket League or Valorant.

## Install

Clone the repository with: `git clone https://github.com/ben-agnew/discordGo.git`
Change directories into the downloaded folder 
Once dowloaded run: `go install`
Create a `.env` file in the directory
Within the `.env` file add the following lines:

    TOKEN="DISCORD_BOT_TOKEN"
    VAL_API="VAL_API_URL"
    RL_API="RL_API_URL"

You will need to supply your own Discord bot token and API URLs. 
A good Valorant API to use is: `https://api.henrikdev.xyz/valorant/v2/mmr/`


## Running the Bot

### Running on Linux
