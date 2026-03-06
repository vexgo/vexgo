import { useState, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { postsApi, categoriesApi, uploadApi } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';
import type { Category } from '@/types';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { RichTextEditor } from '@/components/editor/RichTextEditor';
import {
  Loader2, Save, Send, X, Image as ImageIcon,
  Plus, ArrowLeft
} from 'lucide-react';

export function WritePostPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const { user } = useAuth();
  const isEditMode = !!id;

  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [excerpt, setExcerpt] = useState('');
  const [category, setCategory] = useState('');
  const [tags, setTags] = useState<string[]>([]);
  const [tagInput, setTagInput] = useState('');
  const [coverImage, setCoverImage] = useState('');
  const [categories, setCategories] = useState<Category[]>([]);
  const [saving, setSaving] = useState(false);
  const [uploadingImage, setUploadingImage] = useState(false);

  // 判断用户角色
  const isContributor = user?.role === 'contributor';

  useEffect(() => {
    loadCategories();
    if (isEditMode) {
      loadPost();
    }
  }, [id]);

  const loadCategories = async () => {
    try {
      const response = await categoriesApi.getCategories();
      setCategories(response.data.categories);
    } catch (error) {
      console.error('加载分类失败:', error);
    }
  };

  const loadPost = async () => {
    try {
      const response = await postsApi.getPost(id!);
      const post = response.data.post;
      setTitle(post.title);
      setContent(post.content);
      setExcerpt(post.excerpt || '');
      setCategory(post.category);
      setTags(post.tags || []);
      setCoverImage(post.coverImage || '');
    } catch (error) {
      console.error('加载文章失败:', error);
      navigate('/');
    }
  };

  const handleAddTag = () => {
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setTags([...tags, tagInput.trim()]);
      setTagInput('');
    }
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove));
  };

  const handleCoverImageUpload = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    if (file.size > 10 * 1024 * 1024) {
      alert('图片大小不能超过10MB');
      return;
    }

    setUploadingImage(true);
    try {
      // 按比例压缩图片到封面尺寸上限，保持宽高比
      // 生成固定尺寸的封面图，保持原图完整可见：缩放原图至目标尺寸内并居中，周围填充白色背景（letterbox）
      const resizeImageFile = (file: File, targetWidth = 1200, targetHeight = 630, quality = 0.85) => {
        return new Promise<File>((resolve, reject) => {
          const img = new Image();
          img.onload = () => {
            // 强制拉伸原图到目标尺寸（允许不按比例缩放），以便在封面区域完整显示
            const drawW = targetWidth;
            const drawH = targetHeight;

            // 创建目标画布，填充白色背景以避免透明区域在部分浏览器显示异常
            const canvas = document.createElement('canvas');
            canvas.width = targetWidth;
            canvas.height = targetHeight;
            const ctx = canvas.getContext('2d');
            if (!ctx) {
              reject(new Error('无法获取 canvas 上下文'));
              return;
            }
            // 背景色（可改为透明或其它颜色）
            ctx.fillStyle = '#ffffff';
            ctx.fillRect(0, 0, targetWidth, targetHeight);
            // 直接拉伸绘制到画布上，填满目标尺寸
            ctx.drawImage(img, 0, 0, drawW, drawH);

            canvas.toBlob((blob) => {
              if (!blob) {
                reject(new Error('压缩失败'));
                return;
              }
              const newFile = new File([blob], file.name.replace(/\.[^.]+$/, '.jpg'), { type: 'image/jpeg' });
              resolve(newFile);
            }, 'image/jpeg', quality);
          };
          img.onerror = (err) => reject(err);
          img.src = URL.createObjectURL(file);
        });
      };

      let uploadFile = file;
      try {
        uploadFile = await resizeImageFile(file);
      } catch (err) {
        // 如果压缩失败，回退使用原始文件
        console.warn('图片压缩失败，使用原始文件上传', err);
        uploadFile = file;
      }

      const response = await uploadApi.uploadFile(uploadFile);
      setCoverImage(response.data.file!.url);
    } catch (error) {
      console.error('上传封面图失败:', error);
      alert('上传封面图失败，请重试');
    } finally {
      setUploadingImage(false);
    }
  };

  const handleSubmit = async (status: 'published' | 'draft' | 'pending') => {
    if (!title.trim()) {
      alert('请输入文章标题');
      return;
    }
    if (!content.trim()) {
      alert('请输入文章内容');
      return;
    }
    if (!category) {
      alert('请选择文章分类');
      return;
    }

    setSaving(true);
    try {
      const postData = {
        title: title.trim(),
        content,
        category,
        tags,
        excerpt: excerpt.trim() || content.substring(0, 200) + '...',
        coverImage: coverImage || undefined,
        status: isContributor && status === 'published' ? 'pending' : status
      };

      if (isEditMode) {
        await postsApi.updatePost(id!, postData);
        navigate(`/post/${id}`);
      } else {
        const response = await postsApi.createPost(postData);
        navigate(`/post/${response.data.post.id}`);
      }
    } catch (error) {
      console.error('保存文章失败:', error);
      alert('保存文章失败，请重试');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8 max-w-4xl">
      {/* 头部 */}
      <div className="flex items-center justify-between mb-8">
        <Button variant="ghost" size="sm" onClick={() => navigate(-1)}>
          <ArrowLeft className="w-4 h-4 mr-2" />
          返回
        </Button>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            onClick={() => handleSubmit('draft')}
            disabled={saving}
          >
            <Save className="w-4 h-4 mr-2" />
            保存草稿
          </Button>
          {/* 投稿者只能提交待审核文章，不能直接发布 */}
          {isContributor ? (
            <Button
              onClick={() => handleSubmit('pending')}
              disabled={saving}
            >
              {saving ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Send className="w-4 h-4 mr-2" />
              )}
              提交审核
            </Button>
          ) : (
            <Button
              onClick={() => handleSubmit('published')}
              disabled={saving}
            >
              {saving ? (
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              ) : (
                <Send className="w-4 h-4 mr-2" />
              )}
              {isEditMode ? '更新' : '发布'}
            </Button>
          )}
        </div>
      </div>

      {/* 表单 */}
      <div className="space-y-6">
        {/* 标题 */}
        <div>
          <Input
            placeholder="请输入文章标题..."
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            className="text-2xl font-bold border-0 border-b rounded-none px-0 focus-visible:ring-0"
          />
        </div>

        {/* 封面图 */}
        <Card>
          <CardContent className="p-4">
            <Label className="block mb-2">封面图</Label>
            {coverImage ? (
              <div className="relative">
                <img
                  src={coverImage}
                  alt="封面"
                  className="w-full h-48 object-fill rounded-lg"
                />
                <Button
                  variant="destructive"
                  size="sm"
                  className="absolute top-2 right-2"
                  onClick={() => setCoverImage('')}
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
            ) : (
              <div className="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center">
                <input
                  type="file"
                  accept="image/*"
                  onChange={handleCoverImageUpload}
                  className="hidden"
                  id="cover-image"
                />
                <label
                  htmlFor="cover-image"
                  className="cursor-pointer flex flex-col items-center"
                >
                  {uploadingImage ? (
                    <Loader2 className="w-8 h-8 text-gray-400 animate-spin mb-2" />
                  ) : (
                    <ImageIcon className="w-8 h-8 text-gray-400 mb-2" />
                  )}
                  <span className="text-sm text-gray-500">
                    {uploadingImage ? '上传中...' : '点击上传封面图'}
                  </span>
                  <span className="text-xs text-gray-400 mt-1">
                    支持 JPG、PNG、GIF，最大 10MB
                  </span>
                </label>
              </div>
            )}
          </CardContent>
        </Card>

        {/* 分类和标签 */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* 分类 */}
          <div>
            <Label htmlFor="category" className="block mb-2">分类 *</Label>
            <Select value={category} onValueChange={setCategory}>
              <SelectTrigger>
                <SelectValue placeholder="选择分类" />
              </SelectTrigger>
              <SelectContent>
                {categories.map((cat) => (
                  <SelectItem key={cat.id} value={cat.id}>
                    {cat.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* 标签 */}
          <div>
            <Label htmlFor="tags" className="block mb-2">标签</Label>
            <div className="flex gap-2">
              <Input
                id="tags"
                placeholder="添加标签"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === 'Enter') {
                    e.preventDefault();
                    handleAddTag();
                  }
                }}
              />
              <Button type="button" onClick={handleAddTag} variant="outline">
                <Plus className="w-4 h-4" />
              </Button>
            </div>
          </div>
        </div>

        {/* 标签展示 */}
        {tags.length > 0 && (
          <div className="flex flex-wrap gap-2">
            {tags.map((tag) => (
              <Badge key={tag} variant="secondary" className="flex items-center gap-1">
                {tag}
                <button
                  onClick={() => handleRemoveTag(tag)}
                  className="ml-1 hover:text-destructive"
                >
                  <X className="w-3 h-3" />
                </button>
              </Badge>
            ))}
          </div>
        )}

        {/* 摘要 */}
        <div>
          <Label htmlFor="excerpt" className="block mb-2">摘要</Label>
          <Textarea
            id="excerpt"
            placeholder="请输入文章摘要（可选，不填写将自动提取正文前200字）"
            value={excerpt}
            onChange={(e) => setExcerpt(e.target.value)}
            rows={3}
          />
        </div>

        {/* 富文本编辑器 */}
        <div>
          <Label className="block mb-2">正文内容 *</Label>
          <RichTextEditor
            content={content}
            onChange={setContent}
            placeholder="开始写作..."
          />
        </div>
      </div>
    </div>
  );
}
