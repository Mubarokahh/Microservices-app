pipeline { 
    agent any

    tools {
        go 'go-1.20' // Ensure go-1.20 is installed in Jenkins global tools
    }

    environment {
        GO111MODULE = 'on'
        DOCKER_IMAGE = "results-service"
        KUBECONFIG = credentials('kubeconfig-kind') 
        RELEASE_NAME = "results"
    }

    stages {
        stage('Checkout the results-service Branch') {
            steps {
               // Corrected syntax for git
               //git branch: 'main', credentialsId: 'github-credentials', url: 'https://github.com/AdekunleDally/voting-app.git', timeout:30
               git branch: 'results-service', credentialsId: 'my-github-credentials', url: 'git@github.com:AdekunleDally/voting-app.git'

            }
        }

        stage('Test') {
            steps {
                
                bat 'go test .' // Running Go tests in results-service directory on Windows
                
            }
        }

        stage('Build the results-service Docker Image') {
            steps {
               
                // Using Jenkins 'withCredentials' to handle the .env file securely
                withCredentials([file(credentialsId: 'results-service-env', variable: 'ENV_FILE')]) {

                // Use 'bat' to run Windows commands instead of 'sh'
                bat 'copy %ENV_FILE% .env'  // Windows equivalent of 'cp' command

                // Build the Docker image using the Windows-friendly command
                bat 'docker build -t results-service .'
                }  
            }
        }

        stage('Push the results-service Docker Image to DockerHub') {
            steps {
                script {
                    withDockerRegistry([credentialsId: 'docker-credentials', url: 'https://registry.hub.docker.com']) {
                        bat 'docker tag "results-service" "results-service:latest"'
                        bat 'docker push "results-service:latest"'
                    }
                }
            }
        }
        
        stage('Load image to KIND Cluster') {
            steps {
                bat 'kind load docker-image results-service:latest --name votingapp-microservice'
            }
        }

        stage('Deploy with Helm') {
            steps {
                 bat "helm upgrade --install ${RELEASE_NAME} ./results-chart -f ./results-chart/values.yaml --kubeconfig=${KUBECONFIG} --set image.repository=${DOCKER_IMAGE} --set image.tag=\"latest\""
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