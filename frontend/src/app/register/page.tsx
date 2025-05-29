'use client';

import { useRouter } from 'next/navigation';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import * as z from 'zod';
import { api } from '@/lib/api';
import { toast } from 'react-hot-toast';
import { Label } from '@/components/ui/label';
import { Input } from '@/components/ui/input';
import { AuthFormLayout } from '@/components/ui/AuthFormLayout';

const registerSchema = z
    .object({
        username: z.string().min(1, 'Введите логин'),
        password: z.string().min(6, 'Минимум 6 символов'),
        confirmPassword: z.string().min(6, 'Минимум 6 символов'),
    })
    .refine(data => data.password === data.confirmPassword, {
        message: 'Пароли не совпадают',
        path: ['confirmPassword'],
    });
type RegisterData = z.infer<typeof registerSchema>;

export default function RegisterPage() {
    const router = useRouter();
    const { register: reg, handleSubmit, formState: { errors, isSubmitting } } = useForm<RegisterData>({
        resolver: zodResolver(registerSchema),
    });

    const onSubmit = async (data: RegisterData) => {
        try {
            await api.post('/auth/register', { username: data.username, password: data.password });
            toast.success('Регистрация успешна!');
            router.push('/login');
        } catch (err: any) {
            toast.error(err.response?.data?.message || 'Ошибка регистрации');
        }
    };

    return (
        <AuthFormLayout
            title="Регистрация"
            isSubmitting={isSubmitting}
            submitLabel="Зарегистрироваться"
            submittingLabel="Регистрируем..."
            footerText="Есть аккаунт?"
            footerLink={{ href: '/login', label: 'Войдите' }}
            onSubmit={handleSubmit(onSubmit)}
        >
            <div>
                <Label htmlFor="username">Логин</Label>
                <Input id="username" type="text" placeholder="your_username" {...reg('username')} />
                {errors.username && <p className="text-red-500 text-sm mt-1">{errors.username.message}</p>}
            </div>

            <div>
                <Label htmlFor="password">Пароль</Label>
                <Input id="password" type="password" placeholder="••••••••" {...reg('password')} />
                {errors.password && <p className="text-red-500 text-sm mt-1">{errors.password.message}</p>}
            </div>

            <div>
                <Label htmlFor="confirmPassword">Подтвердите пароль</Label>
                <Input
                    id="confirmPassword"
                    type="password"
                    placeholder="••••••••"
                    {...reg('confirmPassword')}
                />
                {errors.confirmPassword && (
                    <p className="text-red-500 text-sm mt-1">{errors.confirmPassword.message}</p>
                )}
            </div>
        </AuthFormLayout>
    );
}