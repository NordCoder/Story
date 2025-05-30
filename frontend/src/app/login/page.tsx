'use client';
export const dynamic = 'force-dynamic';

import { useRouter, useSearchParams } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { useAuth } from '@/hooks/useAuth';
import { toast } from 'react-hot-toast';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { AuthFormLayout } from '@/components/ui/AuthFormLayout';

const loginSchema = z.object({
    username: z.string().min(1, 'Введите логин'),
    password: z.string().min(6, 'Минимум 6 символов'),
});
type LoginData = z.infer<typeof loginSchema>;

export default function LoginPage() {
    const router = useRouter();
    const params = useSearchParams();
    const { setToken } = useAuth();
    const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<LoginData>({
        resolver: zodResolver(loginSchema),
    });

    const onSubmit = async (data: LoginData) => {
        try {
            const res = await api.post('/auth/login', data);
            setToken(res.data.accessToken);
            toast.success('Успешный вход');
            router.push(params.get('from') || '/');
        } catch (err: any) {
            toast.error(err.response?.data?.message || 'Ошибка входа');
        }
    };

    return (
        <AuthFormLayout
            title="Вход"
            isSubmitting={isSubmitting}
            submitLabel="Войти"
            submittingLabel="Входим..."
            footerText="Нет аккаунта?"
            footerLink={{ href: '/register', label: 'Зарегистрируйтесь' }}
            onSubmit={handleSubmit(onSubmit)}
        >
            <div>
                <Label htmlFor="username">Логин</Label>
                <Input id="username" type="text" placeholder="your_username" {...register('username')} />
                {errors.username && <p className="text-red-500 text-sm mt-1">{errors.username.message}</p>}
            </div>

            <div>
                <Label htmlFor="password">Пароль</Label>
                <Input id="password" type="password" placeholder="••••••••" {...register('password')} />
                {errors.password && <p className="text-red-500 text-sm mt-1">{errors.password.message}</p>}
            </div>
        </AuthFormLayout>
    );
}