# Monarch Launcher (WIP)

### How to use:
1. Download the latest release of [neonite](https://github.com/NeoniteDev/NeoniteV2), then start the server by running the run.bat file.
2. Download the latest release in the [releases tab](https://github.com/fortmods/Monarch/releases) then extract the contents of the .7z to a folder and run the Monarch.exe file.

### About:
- Simple fortnite launcher compatible with private servers on the latest game version (20.10).
- Currently targets neonite, but in the future will be bundled with it's own lobby emulator backend.
- WIP, so obviously not very feature rich at the moment.
- GUI using [wails](https://wails.io)

#### Goals:
- Monarch aims to be an easy to use launcher for as many fortnite builds as possible, while allowing users to easily download a plethora of existing projects through the launcher as well as provide their own local dll files.

#### QnA:
- Q: "Couldnt I just use Carbon?" A: Yes, Carbon works perfectly well for reaching the lobby using neonite, however the goal of this project is to provide more features and a rich user experience greater than that of Carbon. Of course in the current state of Monarch, Carbon has the advantage and is certainly the better option, but as more features pile onto Monarch I intend for this to change!

#### Licensing:
- Monarch: MIT license.
- Wails (Package used to show gui, serves as go replacement to electron): MIT license.
- Go (Programming language Monarch is written in): https://github.com/golang/go/blob/master/LICENSE