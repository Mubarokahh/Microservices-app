pipeline { 
    agent any

    tools {
        go 'go-1.20'
    }

    environment {
        GO111MODULE = 'on'
        DOCKER_IMAGE = "postgres:latest"
        KUBECONFIG = credentials('kubeconfig-kind') 
        RELEASE_NAME = "postgres"
    }

    stages {
        stage('Checkout postgres Code') {
            steps {
               // Clone the repository without specifying a branch or path
               git branch: 'postgres', credentialsId: 'my-github-credentials', url: 'git@github.com/Mubarokahh/Microservices.git' 

            }
        }

        stage('Build the Postgres Image') {
            steps {
                script {
                    bat 'docker build -t "postgres" .'                  
                }
            }
        }

        stage('Push postgres Image') {
            steps {
                script {
                    withDockerRegistry([credentialsId: 'docker-credentials', url: 'https://registry.hub.docker.com']) {
                        bat 'docker tag "postgres" "postgres:latest"'
                        bat 'docker push "postgres:latest"'
                    }
                }
            }
        }

        stage('Load image to KIND Cluster') {
            steps {
                bat 'kind load docker-image postgres:latest --name votingapp-microservice'
            }
        }

        stage('Deploy with Helm') {
            steps {
                bat "helm upgrade --install ${RELEASE_NAME} ./postgres-chart -f ./postgres-chart/values.yaml --kubeconfig=${KUBECONFIG} --set image.repository=${DOCKER_IMAGE} --set image.tag=\"latest\""           
            }
        }

        stage('Test Deployment') {
            steps {
                bat 'kubectl get pods'
                // bat 'kubectl get non-existent-resource'
            }
        }
    }

    post {
        failure {
            script {
                try {
                    // Attempt to rollback the Helm release
                    echo "Deployment failed. Attempting to rollback the release..."
                    def rollbackStatus = bat(script: "helm rollback ${RELEASE_NAME}", returnStatus: true)

                    // Check if rollback was successful based on the return status
                    if (rollbackStatus == 0) {
                        echo "Rollback completed successfully."
                    } else {
                        error("Helm rollback failed with exit code ${rollbackStatus}.")
                    }
                } catch (Exception e) {
                    // Catch any unexpected errors
                    error("An error occurred during the rollback process: ${e.message}")
                }
            }
        }
    }
}