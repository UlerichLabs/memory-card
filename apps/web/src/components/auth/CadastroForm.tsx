import { useState, type FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { authService, AuthApiError } from '@/services/authService'
import { AUTH_API_ERROR_MESSAGES } from './authConstants'
import { validarCadastroForm, type CadastroFormData, type CadastroValidationErrors } from './cadastroValidation'

export function CadastroForm() {
  const navigate = useNavigate()
  const [formData, setFormData] = useState<CadastroFormData>({
    nome: '',
    email: '',
    senha: '',
    confirmacaoSenha: '',
  })
  const [errors, setErrors] = useState<CadastroValidationErrors>({})
  const [apiError, setApiError] = useState<string | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  function handleChange(field: keyof CadastroFormData, value: string) {
    setFormData((prev) => ({ ...prev, [field]: value }))
    if (errors[field]) {
      setErrors((prev) => ({ ...prev, [field]: undefined }))
    }
  }

  async function handleSubmit(e: FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setApiError(null)

    const validationErrors = validarCadastroForm(formData)
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors)
      return
    }

    setIsLoading(true)
    try {
      await authService.cadastrar({
        nome: formData.nome.trim(),
        email: formData.email.trim(),
        senha: formData.senha,
      })
      navigate('/login')
    } catch (err: unknown) {
      if (err instanceof AuthApiError) {
        setApiError(AUTH_API_ERROR_MESSAGES[err.codigo] || AUTH_API_ERROR_MESSAGES.fallback)
      } else {
        setApiError(AUTH_API_ERROR_MESSAGES.fallback)
      }
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4" noValidate>
      <div className="space-y-1.5">
        <Label htmlFor="nome">Nome</Label>
        <Input
          id="nome"
          name="nome"
          type="text"
          value={formData.nome}
          onChange={(e) => handleChange('nome', e.target.value)}
          disabled={isLoading}
          aria-invalid={!!errors.nome}
        />
        {errors.nome && (
          <p className="text-xs text-destructive" role="alert">
            {errors.nome}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="email">Email</Label>
        <Input
          id="email"
          name="email"
          type="email"
          value={formData.email}
          onChange={(e) => handleChange('email', e.target.value)}
          disabled={isLoading}
          aria-invalid={!!errors.email}
        />
        {errors.email && (
          <p className="text-xs text-destructive" role="alert">
            {errors.email}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="senha">Senha</Label>
        <Input
          id="senha"
          name="senha"
          type="password"
          value={formData.senha}
          onChange={(e) => handleChange('senha', e.target.value)}
          disabled={isLoading}
          aria-invalid={!!errors.senha}
        />
        {errors.senha && (
          <p className="text-xs text-destructive" role="alert">
            {errors.senha}
          </p>
        )}
      </div>

      <div className="space-y-1.5">
        <Label htmlFor="confirmacaoSenha">Confirmação de senha</Label>
        <Input
          id="confirmacaoSenha"
          name="confirmacaoSenha"
          type="password"
          value={formData.confirmacaoSenha}
          onChange={(e) => handleChange('confirmacaoSenha', e.target.value)}
          disabled={isLoading}
          aria-invalid={!!errors.confirmacaoSenha}
        />
        {errors.confirmacaoSenha && (
          <p className="text-xs text-destructive" role="alert">
            {errors.confirmacaoSenha}
          </p>
        )}
      </div>

      {apiError && (
        <p className="text-xs text-destructive" role="alert">
          {apiError}
        </p>
      )}

      <Button type="submit" disabled={isLoading} className="w-full">
        {isLoading ? 'Cadastrando...' : 'Cadastrar'}
      </Button>
    </form>
  )
}
