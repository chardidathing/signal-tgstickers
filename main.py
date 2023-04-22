import requests
from io import BytesIO
from PIL import Image
import os
from configparser import ConfigParser

# read the configuration file
config = ConfigParser()
config.read('config.ini')

# read the bot token and output directory from the configuration file
bot_token = config['CONFIG']['bot_token']
output_dir = config['CONFIG']['output_dir']

# define the desired maximum size for the images (shouldn't need to change this if you're using for signal)
max_size = (512, 512)

# define the desired output formats - png, webp, etc. 
# if using multiple, separate like ['webp', 'png']
output_formats = ['webp']

# prompt the user for the sticker pack name and strip the full url if it exists
sticker_pack_name = input('Enter the URL/name of the Telegram sticker pack: ')
sticker_pack_name = sticker_pack_name.replace('https://t.me/addstickers/', '')

# make a request to the telegram bot api to get the sticker pack
response = requests.get(f'https://api.telegram.org/bot{bot_token}/getStickerSet?name={sticker_pack_name}')

# parse the response to get the list of stickers
stickers = response.json()['result']['stickers']

# loop through each sticker and convert it
print(f'Converting {len(stickers)} stickers...')
for sticker in stickers:
    # download the image
    file_id = sticker['file_id']
    response = requests.get(f'https://api.telegram.org/bot{bot_token}/getFile?file_id={file_id}')
    file_path = response.json()['result']['file_path']
    response = requests.get(f'https://api.telegram.org/file/bot{bot_token}/{file_path}')
    image = Image.open(BytesIO(response.content))
    
    # resize the image if necessary
    if image.size[0] > max_size[0] or image.size[1] > max_size[1]:
        print(f'Resizing {sticker["file_unique_id"]}...')
        image.thumbnail(max_size)
    
    # convert the image to the desired formats
    for format in output_formats:
        output_path = os.path.join(output_dir, f'{sticker["file_unique_id"]}.{format}')
        print(f'Converting {sticker["file_unique_id"]} to {format}...')
        image.save(output_path, format=format)

print('Conversion complete!')
