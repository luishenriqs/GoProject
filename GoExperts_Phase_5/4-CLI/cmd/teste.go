/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/

/*
Arquivo responsável por definir e registrar o comando "teste" da aplicação CLI.

Este comando demonstra a estrutura básica de um subcomando com Cobra,
incluindo definição do comando, descrição, leitura de flags e execução
de lógica condicional no terminal.

Comportamento atual:
- registra o subcomando "teste" no comando raiz;
- expõe a flag local "--comando" com atalho "-c";
- imprime "ping" quando o valor informado for "ping";
- imprime "pong" para qualquer outro valor.

Exemplos de uso:
- go run main.go teste
- go run main.go teste --comando ping
- go run main.go teste -c ping
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// testeCmd represents the teste command
var testeCmd = &cobra.Command{
	Use:   "teste",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		comando, _ := cmd.Flags().GetString("comando")
		if comando == "ping" {
			cmd.Println("ping")
		} else {
			cmd.Println("pong")
		}
	},
}

func init() {
	rootCmd.AddCommand(testeCmd)
	testeCmd.Flags().StringP("comando", "c", "", "Escolha ping ou pong")
}
