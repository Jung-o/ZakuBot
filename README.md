# ZakuBot

This is an anime discord bot. You can collect anime characters and trade them with your friends.

## Description

It uses a MongoDB database to access characters. It's based on the discordgo library.

## Getting Started

### Dependencies
* The bot needs a local.env file with your discord bot token.
```
BOT_TOKEN=<YOUR_DISCORD_BOT_TOKEN>
```

#### MongoDB (localhost:27017)
The Program needs two collections already filled.

##### Characters
```
characterId: String
name: String
series: String
owners: Array
```

##### Artworks
```
charName: String
characterId: String
artworkId: Int32
```

### Executing program
* Download all the dependencies
```
go get -d ./...
```

* Then run the main.go file
```
go run main.go
```


## License

This project is licensed under the GNU General Public License v3.0 License - see the LICENSE.txt file for details
