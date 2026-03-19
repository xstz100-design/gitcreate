from io import BytesIO
from pathlib import Path
from typing import Optional
from PIL import Image
import hashlib


class ImageService:
    """图片处理服务 - 生成缩略图，提升移动端加载速度"""
    
    def __init__(self, upload_dir: str = "uploads"):
        self.upload_dir = Path(upload_dir)
        self.upload_dir.mkdir(exist_ok=True)
        self.thumbnail_dir = self.upload_dir / "thumbnails"
        self.thumbnail_dir.mkdir(exist_ok=True)
    
    def create_thumbnail(
        self,
        image_path: str,
        size: tuple[int, int] = (200, 200),
        quality: int = 85
    ) -> Optional[str]:
        """
        创建缩略图
        
        Args:
            image_path: 原图路径
            size: 缩略图尺寸 (宽, 高)
            quality: 压缩质量 1-100
            
        Returns:
            缩略图路径或 None
        """
        try:
            # 打开原图
            with Image.open(image_path) as img:
                # 转换为 RGB (处理 PNG 透明通道)
                if img.mode in ('RGBA', 'LA', 'P'):
                    background = Image.new('RGB', img.size, (255, 255, 255))
                    if img.mode == 'P':
                        img = img.convert('RGBA')
                    background.paste(img, mask=img.split()[-1] if img.mode == 'RGBA' else None)
                    img = background
                
                # 保持宽高比缩放
                img.thumbnail(size, Image.Resampling.LANCZOS)
                
                # 生成缩略图文件名
                original_name = Path(image_path).stem
                thumbnail_name = f"{original_name}_thumb.jpg"
                thumbnail_path = self.thumbnail_dir / thumbnail_name
                
                # 保存缩略图
                img.save(thumbnail_path, "JPEG", quality=quality, optimize=True)
                
                return str(thumbnail_path)
        
        except Exception as e:
            print(f"Failed to create thumbnail: {e}")
            return None
    
    def optimize_image(
        self,
        image_path: str,
        max_size: tuple[int, int] = (1200, 1200),
        quality: int = 90
    ) -> Optional[str]:
        """
        优化图片大小
        
        Args:
            image_path: 图片路径
            max_size: 最大尺寸
            quality: 压缩质量
            
        Returns:
            优化后的图片路径
        """
        try:
            with Image.open(image_path) as img:
                # 如果图片过大，缩小
                if img.width > max_size[0] or img.height > max_size[1]:
                    img.thumbnail(max_size, Image.Resampling.LANCZOS)
                
                # 转换为 RGB
                if img.mode in ('RGBA', 'LA', 'P'):
                    background = Image.new('RGB', img.size, (255, 255, 255))
                    if img.mode == 'P':
                        img = img.convert('RGBA')
                    background.paste(img, mask=img.split()[-1] if img.mode == 'RGBA' else None)
                    img = background
                
                # 保存优化后的图片
                img.save(image_path, "JPEG", quality=quality, optimize=True)
                
                return image_path
        
        except Exception as e:
            print(f"Failed to optimize image: {e}")
            return None
    
    @staticmethod
    def get_image_hash(image_data: bytes) -> str:
        """生成图片哈希值，用于去重"""
        return hashlib.md5(image_data).hexdigest()


# 全局实例
image_service = ImageService()
