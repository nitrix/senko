# List of commands

## Autojoin module

- `/autojoin on <channel>` Enables automatically joining/leaving to designated channel.
- `/autojoin off` Turns of When turned on, it will automatically join and leave the designated channel when someone is present.

## Core module

The basic operational commands to manipulate the bot around.

- `/say <text>` Says something in the current voice channel.
- `/join <channel>` Joins a voice channel by name. Defaults to your current channel when omitted.
- `/leave` Leaves the current voice channel.

## Deejay

This module is clever enough to normalize the audio of songs and not destroy your ears.

Its playback volume can be independently controlled per guild from the rest of the usual voice communications.

- `/play <what>` Queues up a song (by name or url) to play on the current voice channel.
- `/pause` Pauses playback; can be resumed.
- `/volume` Modifies the playback volume as a percentage (between 0% and 100%).
- `/resume` Resumes the playback where it left off.
- `/stop` Stops playback and clears the queue.

## Jarvis module

Turns the voice commands into text commands.