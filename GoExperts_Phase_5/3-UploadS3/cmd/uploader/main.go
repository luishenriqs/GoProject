package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	maxConcurrentUploads = 100
	maxRetryAttempts     = 3
	uploadTimeout        = 30 * time.Second
)

type uploadTask struct {
	filename string
	attempt  int
}

var (
	s3Client *s3.Client
	s3Bucket string
)

/*
loadAWSConfig carrega a configuração da AWS SDK v2 usando credenciais
hardcoded para uso estritamente local no contexto do curso.

Parâmetros:
- ctx: contexto base para carregamento da configuração.

Retorno:
- *s3.Client: cliente S3 inicializado.
- string: nome do bucket configurado.
- error: erro de configuração, quando houver.
*/
func loadAWSConfig(ctx context.Context) (*s3.Client, string, error) {
	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("us-east-1"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				"YOUR_ACCESS_KEY",
				"YOUR_SECRET_KEY",
				"",
			),
		),
	)
	if err != nil {
		return nil, "", fmt.Errorf("erro ao carregar configuração AWS: %w", err)
	}

	client := s3.NewFromConfig(cfg)
	bucket := "goexpert-lhp-bucket"

	return client, bucket, nil
}

/*
processUpload executa o upload de um único arquivo para o S3.

Parâmetros:
- filename: nome do arquivo dentro de ./tmp.

Retorno:
- error: erro de abertura do arquivo ou de upload, quando houver.
*/
func processUpload(filename string) error {
	completeFileName := fmt.Sprintf("./tmp/%s", filename)
	// fmt.Printf("Uploading file %s to bucket %s\n", completeFileName, s3Bucket)

	f, err := os.Open(completeFileName)
	if err != nil {
		return fmt.Errorf("erro ao abrir arquivo %s: %w", completeFileName, err)
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(context.Background(), uploadTimeout)
	defer cancel()

	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s3Bucket,
		Key:    &filename,
		Body:   f,
	})
	if err != nil {
		return fmt.Errorf("erro ao enviar arquivo %s: %w", completeFileName, err)
	}

	// fmt.Printf("File %s uploaded successfully\n", completeFileName)

	return nil
}

/*
uploadWorker consome tarefas de upload do canal jobs, executa o upload e,
em caso de falha, reenfileira a tarefa até o limite definido de tentativas.

Parâmetros:
- id: identificador do worker para rastreabilidade.
- jobs: canal de entrada de tarefas.
- wg: wait group responsável por contabilizar cada tarefa enfileirada.
*/
func uploadWorker(id int, jobs chan uploadTask, wg *sync.WaitGroup) {
	for task := range jobs {
		err := processUpload(task.filename)
		if err != nil {
			// fmt.Printf("Worker %d: %v\n", id, err)

			if task.attempt < maxRetryAttempts {
				fmt.Printf(
					"Worker %d: retrying file %s (attempt %d of %d)\n",
					id,
					task.filename,
					task.attempt+1,
					maxRetryAttempts,
				)

				wg.Add(1)
				jobs <- uploadTask{
					filename: task.filename,
					attempt:  task.attempt + 1,
				}
			} else {
				fmt.Printf(
					"Worker %d: max retries reached for file %s\n",
					id,
					task.filename,
				)
			}
		}

		wg.Done()
	}
}

/*
main inicializa o client S3, percorre os arquivos do diretório ./tmp,
cria workers concorrentes para upload e aguarda a conclusão de todas as tarefas.

Parâmetros:
- nenhum.

Retorno:
- nenhum.
*/
func main() {
	var err error

	s3Client, s3Bucket, err = loadAWSConfig(context.Background())
	if err != nil {
		panic(err)
	}

	dir, err := os.Open("./tmp")
	if err != nil {
		panic(err)
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		panic(err)
	}

	jobs := make(chan uploadTask, maxConcurrentUploads)
	var wg sync.WaitGroup

	for i := 0; i < maxConcurrentUploads; i++ {
		go uploadWorker(i+1, jobs, &wg)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		wg.Add(1)
		jobs <- uploadTask{
			filename: file.Name(),
			attempt:  1,
		}
	}

	wg.Wait()
	close(jobs)
}
