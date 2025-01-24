pipeline { 
    agent any

    tools {
        go 'go-1.20'
    }

    environment {
        GO111MODULE = 'on'
        KUBECONFIG = credentials('kubeconfig-kind') // Using kubeconfig
        DOCKER_IMAGE = "redis:latest"
        RELEASE_NAME = "redis"
    }

    stages {
        stage('Checkout Redis Code') {
            steps {
               // Clone the repository without specifying a branch or path
               git branch: 'redis', credentialsId: 'my-github-credentials', url: 'git@github.com/Mubarokahh/voting-app.git' 

            }
        }

        stage('Build redis Image') {
            steps {
                script {
                    bat 'docker build -t "redis" .'
                }
            }
        }

        stage('Push Redis Image') {
            steps {
                script {
                    withDockerRegistry([credentialsId: 'docker-credentials', url: 'https://registry.hub.docker.com']) {
                        bat 'docker tag "redis" "redis:latest"'
                        bat 'docker push "/redis:latest"'
                    }
                }
            }
        }

        stage('Load image to KIND Cluster') {
            steps {
                bat 'kind load docker-imageredis:latest --name votingapp-microservice'
            }
        }

        stage('Deploy with Helm') {
            steps {
                bat "helm upgrade --install ${RELEASE_NAME} ./redis-chart -f ./redis-chart/values.yaml --kubeconfig=${KUBECONFIG} --set image.repository=${DOCKER_IMAGE} --set image.tag=\"latest\""     
            }
        }

        stage('Test Deployment') {
            steps {
                bat 'kubectl get pods'
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