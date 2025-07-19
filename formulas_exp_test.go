package t

import (
	"fmt"
	"github.com/expr-lang/expr"
	"strings"
	"testing"
)

func TestExprFormulaEvaluationExp3(t *testing.T) {
	// 1. Expressão com variáveis
	formula := "a + b * c > 10"

	// 2. Ambiente de dados (as variáveis que a fórmula usará)
	env := map[string]any{
		"a": 5,
		"b": 2,
		"c": 3,
	}

	// 3. Compilar a expressão
	// O `Compile` transforma a string da fórmula em um programa executável.
	// É uma boa prática compilar uma vez e reusar o programa se a fórmula for constante.
	program, err := expr.Compile(formula, expr.Env(env))
	if err != nil {
		fmt.Printf("Erro ao compilar a fórmula: %v\n", err)
		return
	}

	// 4. Executar o programa com o ambiente
	output, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula: %v\n", err)
		return
	}

	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Ambiente: %v\n", env)
	fmt.Printf("Resultado: %v (Tipo: %T)\n", output, output)

	fmt.Println("---")

	// 5. Reutilizando o programa com um ambiente diferente
	env2 := map[string]any{
		"a": 1,
		"b": 20,
		"c": 0.5,
	}
	output2, err := expr.Run(program, env2)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula com env2: %v\n", err)
		return
	}
	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Ambiente: %v\n", env2)
	fmt.Printf("Resultado com novo ambiente: %v (Tipo: %T)\n", output2, output2)
}

type User struct {
	Name string
	Age  int
	Role string
}

func TestExprFormulaEvaluationExp4(t *testing.T) {
	// 1. Expressão com acesso a campos de struct e chamadas de função
	// Note o uso de `strings.ToUpper` e `len`
	formula := `user.Age >= 18 and user.Role == "admin" or len(user.Name) > 5 and strings.ToUpper(user.Name) != "BOB"`

	// 2. Ambiente de dados, incluindo uma struct e uma função externa
	env := map[string]any{
		"user": User{Name: "Alice", Age: 25, Role: "admin"},
		// Adicionando a função strings.ToUpper ao ambiente
		"strings": map[string]any{
			"ToUpper": strings.ToUpper,
		},
	}

	// 3. Compilar a expressão
	program, err := expr.Compile(formula, expr.Env(env))
	if err != nil {
		fmt.Printf("Erro ao compilar a fórmula: %v\n", err)
		return
	}

	// 4. Executar o programa
	output, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula: %v\n", err)
		return
	}

	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Ambiente: %v\n", env)
	fmt.Printf("Resultado: %v (Tipo: %T)\n", output, output)

	fmt.Println("---")

	// 5. Teste com outro usuário
	env2 := map[string]any{
		"user": User{Name: "Bob", Age: 17, Role: "user"},
		"strings": map[string]any{
			"ToUpper": strings.ToUpper,
		},
	}
	output2, err := expr.Run(program, env2)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula com env2: %v\n", err)
		return
	}
	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Ambiente: %v\n", env2)
	fmt.Printf("Resultado com Bob: %v (Tipo: %T)\n", output2, output2)
}

func TestExprFormulaEvaluationExp5(t *testing.T) {
	// 1. Expressão com operador ternário e variável potencialmente indefinida
	// Se 'name' for nil ou vazio, retorna "Hello, world!", senão "Hello, [name]!"
	formula := `name == nil || name == "" ? "Hello, world!" : sprintf("Hello, %v!", name)`

	// 2. Ambiente com uma função Go nativa (`fmt.Sprintf`)
	env := map[string]any{
		"sprintf": fmt.Sprintf,
	}

	// 3. Compilar a expressão, permitindo variáveis indefinidas
	// `expr.AllowUndefinedVariables()` é útil quando você espera que algumas variáveis
	// possam não estar presentes no ambiente.
	program, err := expr.Compile(formula, expr.Env(env), expr.AllowUndefinedVariables())
	if err != nil {
		fmt.Printf("Erro ao compilar a fórmula: %v\n", err)
		return
	}

	// 4. Executar com 'name' indefinido
	output1, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula (sem nome): %v\n", err)
		return
	}
	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Resultado (sem nome): %v\n", output1)

	fmt.Println("---")

	// 5. Executar com 'name' definido
	env["name"] = "Go Developer"
	output2, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula (com nome): %v\n", err)
		return
	}
	fmt.Printf("Resultado (com nome): %v\n", output2)
}

func TestExprFormulaEvaluationExp6(t *testing.T) {
	// 1. A fórmula mais simples: Comprimento * Largura
	formula := "comprimento * largura"

	// 2. Definindo os valores no ambiente
	env := map[string]any{
		"comprimento": 10.0, // Em metros, por exemplo
		"largura":     5.0,  // Em metros
	}

	// 3. Compilar a expressão
	// Sempre compile a fórmula uma vez se ela não mudar.
	program, err := expr.Compile(formula, expr.Env(env))
	if err != nil {
		fmt.Printf("Erro ao compilar a fórmula: %v\n", err)
		return
	}

	// 4. Executar o programa com o ambiente
	area, err := expr.Run(program, env)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula: %v\n", err)
		return
	}

	fmt.Printf("Fórmula: \"%s\"\n", formula)
	fmt.Printf("Comprimento: %v, Largura: %v\n", env["comprimento"], env["largura"])
	fmt.Printf("Área calculada: %v (Tipo: %T)\n", area, area)

	fmt.Println("---")

	// 5. Testar com outros valores
	env2 := map[string]any{
		"comprimento": 7.5,
		"largura":     2.0,
	}
	area2, err := expr.Run(program, env2)
	if err != nil {
		fmt.Printf("Erro ao executar a fórmula com env2: %v\n", err)
		return
	}
	fmt.Printf("Comprimento: %v, Largura: %v\n", env2["comprimento"], env2["largura"])
	fmt.Printf("Nova área calculada: %v (Tipo: %T)\n", area2, area2)
}
