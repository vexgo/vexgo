import { useEffect, useState } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { authApi } from '@/lib/api';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Loader2, CheckCircle, XCircle } from 'lucide-react';

export function VerifyEmailPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');
  
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  const [message, setMessage] = useState('');

  const verifyEmail = async (tokenValue: string) => {
    try {
      const response = await authApi.verifyEmail(tokenValue);
      setStatus('success');
      setMessage(response.data.message || '邮箱验证成功！');
    } catch (err: any) {
      setStatus('error');
      setMessage(err.response?.data?.error || '验证失败，请检查链接是否正确');
    }
  };

  useEffect(() => {
    if (!token) {
      setStatus('error');
      setMessage('验证令牌不能为空');
      return;
    }

    verifyEmail(token);
  }, [token]);

  return (
    <div className="container mx-auto px-4 py-16 flex justify-center">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl flex items-center justify-center gap-2">
            {status === 'loading' && <Loader2 className="w-6 h-6 animate-spin" />}
            {status === 'success' && <CheckCircle className="w-6 h-6 text-green-500" />}
            {status === 'error' && <XCircle className="w-6 h-6 text-red-500" />}
            邮箱验证
          </CardTitle>
          <CardDescription>
            {status === 'loading' && '正在验证您的邮箱...'}
            {status === 'success' && '验证完成'}
            {status === 'error' && '验证失败'}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-muted-foreground">{message}</p>
          
          <div className="flex gap-3">
            <Button asChild className="flex-1">
              <Link to="/login">
                前往登录
              </Link>
            </Button>
            {status === 'error' && (
              <Button variant="outline" asChild className="flex-1">
                <Link to="/">
                  返回首页
                </Link>
              </Button>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
