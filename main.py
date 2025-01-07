import os
from telegram import Bot
import asyncio
from typing import Optional
from dotenv import load_dotenv
import subprocess
from pathlib import Path
import shutil

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
            
            converter = TGSConverter()
            
            # Download each sticker
            for i, sticker in enumerate(sticker_set.stickers, 1):
                file = await self.bot.get_file(sticker.file_id)
                
                if sticker.is_animated:
                    extension = ".tgs"
                    output_path = os.path.join(pack_dir, f"sticker_{i}{extension}")
                    await file.download_to_drive(output_path)
                    
                    # Convert TGS to APNG
                    try:
                        apng_path = converter.convert_tgs_to_apng(Path(output_path))
                        print(f"Converted sticker {i} to APNG: {apng_path}")
                    except Exception as e:
                        print(f"Failed to convert sticker {i}: {e}")
                        
                else:
                    # Handle other sticker types as before
                    extension = ".webm" if sticker.is_video else ".webp"
                    output_path = os.path.join(pack_dir, f"sticker_{i}{extension}")
                    await file.download_to_drive(output_path)
                
                print(f"Downloaded sticker {i}/{len(sticker_set.stickers)}")

            print(f"Successfully downloaded sticker pack to: {pack_dir}")
            return pack_dir
            
        except Exception as e:
            print(f"Error downloading sticker pack: {e}")
            return None

class TGSConverter:
    def __init__(self, resolution: int = 256):
        self.resolution = resolution
        if not shutil.which('ffmpeg'):
            raise RuntimeError("ffmpeg is required but not found")

    def convert_tgs_to_apng(self, tgs_file: Path) -> Path:
        """
        Convert a .tgs file to .apng format using lottie_convert.py and ffmpeg
        
        Args:
            tgs_file: Path to the .tgs file
            
        Returns:
            Path to the converted .apng file
        """
        # Create intermediate gif path
        gif_path = tgs_file.with_suffix('.gif')
        apng_path = tgs_file.with_suffix('.apng')
        
        try:
            # Convert TGS to GIF using lottie_convert.py
            subprocess.run([
                './python-lottie/bin/lottie_convert.py',
                str(tgs_file),
                str(gif_path)
            ], check=True)
            
            # Convert GIF to APNG using ffmpeg
            subprocess.run([
                'ffmpeg', '-i', str(gif_path),
                '-plays', '0',  # Infinite loop
                '-vf', f'scale={self.resolution}:-1:flags=lanczos',
                '-f', 'apng',
                str(apng_path)
            ], check=True)
            
            # Clean up intermediate GIF file
            gif_path.unlink()
            
            return apng_path
            
        except subprocess.CalledProcessError as e:
            print(f"Conversion failed: {e}")
            # Clean up any intermediate files
            if gif_path.exists():
                gif_path.unlink()
            if apng_path.exists():
                apng_path.unlink()
            raise

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