import os
from telegram import Bot
import asyncio
from typing import Optional
from dotenv import load_dotenv

class StickerPackDownloader:
    def __init__(self, bot_token: str):
        self.bot = Bot(token=bot_token)
        
    async def download_sticker_pack(self, pack_name: str, output_dir: str = "downloaded_stickers"):
        """
        Downloads all stickers from a specified sticker pack.
        
        Args:
            pack_name: Name of the sticker pack (e.g., 'packname' from 't.me/addstickers/packname')
            output_dir: Directory where stickers will be saved
        """
        try:
            # Get sticker set information
            sticker_set = await self.bot.get_sticker_set(pack_name)
            
            # Create output directory with pack name
            pack_dir = os.path.join(output_dir, f"{sticker_set.name}_{sticker_set.title}")
            os.makedirs(pack_dir, exist_ok=True)
            
            print(f"Downloading sticker pack: {sticker_set.title}")
            print(f"Total stickers: {len(sticker_set.stickers)}")
            
            # Download each sticker
            for i, sticker in enumerate(sticker_set.stickers, 1):
                # Get file path for sticker
                file = await self.bot.get_file(sticker.file_id)
                
                # Determine file extension based on sticker type
                if sticker.is_animated:
                    extension = ".tgs"  # TODO: handle this
                    print("Sticker is animated - .tgs support is not implemented yet")
                elif sticker.is_video:
                    extension = ".webm"  # Video
                else:
                    extension = ".webp"  # Image
                    
                output_path = os.path.join(pack_dir, f"sticker_{i}{extension}")
                
                # Download the file
                await file.download_to_drive(output_path)
                print(f"Downloaded sticker {i}/{len(sticker_set.stickers)}")
                
            print(f"Successfully downloaded sticker pack to: {pack_dir}")
            return pack_dir
            
        except Exception as e:
            print(f"Error downloading sticker pack: {e}")
            return None

async def main():
    # Load environment variables
    load_dotenv()
    
    # Get bot token from environment
    bot_token = os.getenv('TELEGRAM_BOT_TOKEN')
    if not bot_token:
        print("Error: TELEGRAM_BOT_TOKEN not found in .env file")
        return
    
    # Initialize downloader
    downloader = StickerPackDownloader(bot_token)
    
    # Prompt user for pack name
    print("Enter the sticker pack name (the part after 't.me/addstickers/')")
    pack_name = input("Pack name: ").strip()
    
    # Download the sticker pack
    await downloader.download_sticker_pack(pack_name)

if __name__ == "__main__":
    asyncio.run(main())