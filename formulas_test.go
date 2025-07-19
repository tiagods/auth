package t_test

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"testing"
)

func TestFormulaEvaluation(t *testing.T) {
	// Example of evaluating a formula using govaluate
	// This is a simple example, you can replace it with your actual formula logic
	// The formula

	params := map[string]interface{}{
		"valor_transacao": 1000.0,
		"parcelas":        2,
		"juros":           0.03,
	}

	expression, _ := govaluate.NewEvaluableExpression("valor_transacao * juros + parcelas")
	result, _ := expression.Evaluate(params)

	fmt.Println("Resultado:", result) // Resultado: 32.0

}

func TestFormulaEvaluation2(t *testing.T) {
	// Fórmula dinâmica vinda do banco
	formula := "IF(parcelas > 6, valor_transacao * 0.04, valor_transacao * 0.03 + 2)"

	// Função personalizada IF(condição, valor_verdadeiro, valor_falso)
	customFunctions := map[string]govaluate.ExpressionFunction{
		"IF": func(args ...interface{}) (interface{}, error) {
			condition := args[0].(bool)
			if condition {
				return args[1], nil
			}
			return args[2], nil
		},
	}

	// Criar expressão com função personalizada
	expression, err := govaluate.NewEvaluableExpressionWithFunctions(formula, customFunctions)
	if err != nil {
		panic(err)
	}

	// Simular os dados da transação
	parameters := map[string]interface{}{
		"valor_transacao": 1000.0,
		"parcelas":        5,
	}

	result, err := expression.Evaluate(parameters)
	if err != nil {
		panic(err)
	}

	fmt.Printf("🧾 Taxa aplicada: %.2f\n", result)

}
