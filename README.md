# Senko-san

Looks like someone's in need of good pamperin'

![senko](docs/senko.jpg)

Senko-san is the Discord server caretaker my friends and I are using.

It cannot do a whole lot currently and is written quite poorly, but as our needs grow, so will the bot.

## Using Senko

To use the bot, you may:

1. Decide to use the public one that I'm already running, or
2. Decide to host your own private one.

### Public one

Simply click [on this OAuth link](https://discordapp.com/api/oauth2/authorize?client_id=348235222615195662&permissions=51200&scope=bot) and follow the steps.

### Private one

To host Senko on your own infrastructure, you'll need to:

1. **Setup** the application
2. To **build** it and
3. To **run** it.

For the setup, you can [create a new application](https://discordapp.com/developers/applications) on Discord. Then, generate a OAuth2 URL and to visit that URL yourself to invite the bot on your server.

For the [building](https://github.com/nitrix/senko#building-senko) and [running](https://github.com/nitrix/senko#running-senko) steps, refer to their respective sections below.

## Building Senko

### Via Docker

The easiest way to build the project is using Docker to generate an image:

    docker build -t senko .

Where `senko` here refers to the name of the image that'll be created and the `.` is the location of the
`Dockerfile` configuration (found at the root of the repository).

### Manually

It is also possible to build the project manually:

    go build

That should generate a binary named `senko` for the platform you're currently on. If instead you'd prefer to
cross-compile for a different platform, refer to the [official documentation](https://golang.org/doc/install/source#environment) on Golang's website.

## Running Senko

Running the bot should be as simple as launching the executable or 

An environment variable named `DISCORD_TOKEN` or a file of the same name in the current working directory will be used for the authentication.

## Documentation

* [List of commands](docs/commands.md)

## Licensing

This is free and unencumbered software released into the public domain. See the [UNLICENSE](UNLICENSE) file for more details.