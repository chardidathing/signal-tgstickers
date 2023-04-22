# Telegram Sticker Conversion Tool for Signal *(or for anything else???)*

I wanted to pull some of the stickers I use on Telegram and move them into Signal, but didn't want to go through the process of doing it manually, so I took 10x the time in writing a script to do it! Woo,,,,, development.

## Getting Started

#### Requirements
- Python3
  - You will also need `requests` and `pillow`, you can use `pip3 install -r requirements.txt` to install these.
- Telegram Bot Token
  - You can send `/newbot` to `@botfather`, give it a name, and it will create your bot, and it will give you a bot token to use.
- Sticker Pack URL
  - You can get this by selecting the sticker pack you want to download/convert, and clicking the Share icon, and `Copy Link`.

#### Setup
1. Clone/Download the repository.
2. Create a copy of `config.ini.example` called `config.ini`, specify your Bot Token here, and the default output directory of `./out` should be fine, however change it if you like.
3. Run `python3 ./main.py`, you will be prompted for the Sticker Pack URL, and it should begin downloading.

#### Troubleshooting
I haven't had any issues to really add here, but if you have any issues, feel free to let me know by opening an issue and I try to help!

#### Thanks!
If you used the script and thought it was helpful, feel free to star it!
