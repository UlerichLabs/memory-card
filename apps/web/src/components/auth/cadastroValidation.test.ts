import { describe, it, expect } from 'vitest'
import { validarCadastroForm } from './cadastroValidation'

describe('validarCadastroForm', () => {
  const validData = {
    nome: 'Lucas',
    email: 'lucas@example.com',
    senha: 'SenhaForte@123',
    confirmacaoSenha: 'SenhaForte@123',
  }

  it('retorna nenhum erro para dados válidos', () => {
    const errors = validarCadastroForm(validData)
    expect(Object.keys(errors)).toHaveLength(0)
  })

  it('valida obrigatoriedade do nome', () => {
    const errors = validarCadastroForm({ ...validData, nome: '   ' })
    expect(errors.nome).toBe('O nome é obrigatório.')
  })

  it('valida formato do email', () => {
    const invalidEmails = ['', 'lucas', 'lucas@', 'lucas@example', 'lucas @example.com']
    for (const email of invalidEmails) {
      const errors = validarCadastroForm({ ...validData, email })
      expect(errors.email).toBe('Informe um email válido.')
    }
  })

  it('valida critérios de força da senha', () => {
    const weakPasswords = [
      'Ab1!xyz',
      'senhaforte@123',
      'SenhaForte@abc',
      'SenhaForte1234',
    ]
    for (const senha of weakPasswords) {
      const errors = validarCadastroForm({ ...validData, senha, confirmacaoSenha: senha })
      expect(errors.senha).toBe(
        'A senha deve ter no mínimo 8 caracteres, incluindo maiúscula, número e caractere especial.'
      )
    }
  })

  it('valida divergência entre senha e confirmação de senha', () => {
    const errors = validarCadastroForm({
      ...validData,
      confirmacaoSenha: 'OutraSenha@123',
    })
    expect(errors.confirmacaoSenha).toBe('Senha e confirmação não conferem.')
  })
})
