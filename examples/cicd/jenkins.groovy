// Jenkins Pipeline: StackEye Monitoring Integration
// ==================================================
//
// This pipeline sets up monitoring for your application after deployment.
// Add this to your Jenkinsfile or create a shared library.
//
// Required Credentials:
//   stackeye-api-key - Secret text credential containing your API key
//
// Optional Credentials:
//   slack-webhook-url - Secret text for Slack notifications

pipeline {
    agent any

    environment {
        APP_NAME = 'my-application'
        APP_URL = 'https://api.example.com'
        STACKEYE_API_KEY = credentials('stackeye-api-key')
    }

    stages {
        stage('Build') {
            steps {
                echo 'Building application...'
                // Your build steps here
            }
        }

        stage('Test') {
            steps {
                echo 'Running tests...'
                // Your test steps here
            }
        }

        stage('Deploy') {
            steps {
                echo 'Deploying application...'
                // Your deployment steps here
            }
        }

        stage('Setup Monitoring') {
            steps {
                script {
                    // Install StackEye CLI
                    sh '''
                        curl -fsSL https://get.stackeye.io/cli | bash
                        export PATH="$PATH:$HOME/.local/bin"
                        stackeye version
                    '''

                    // Configure StackEye
                    sh '''
                        export PATH="$PATH:$HOME/.local/bin"
                        stackeye setup --no-input
                        stackeye whoami
                    '''

                    // Create probe if not exists
                    sh '''
                        export PATH="$PATH:$HOME/.local/bin"

                        EXISTING_PROBE=$(stackeye probe list -o json | jq -r ".[] | select(.name == \\"${APP_NAME} - Health\\") | .id")

                        if [ -z "$EXISTING_PROBE" ]; then
                            echo "Creating new probe..."
                            stackeye probe create \
                                --name "${APP_NAME} - Health" \
                                --url "${APP_URL}/health" \
                                --interval 60 \
                                --timeout 10 \
                                --json-path-check "$.status" \
                                --json-path-expected "ok"
                        else
                            echo "Probe already exists: $EXISTING_PROBE"
                        fi
                    '''

                    // Run health check
                    sh '''
                        export PATH="$PATH:$HOME/.local/bin"

                        PROBE_ID=$(stackeye probe list -o json | jq -r ".[] | select(.name | contains(\\"${APP_NAME}\\")) | .id" | head -1)

                        if [ -n "$PROBE_ID" ]; then
                            stackeye probe test "$PROBE_ID"
                        fi
                    '''
                }
            }
        }

        stage('Verify Deployment') {
            steps {
                script {
                    // Wait for initial checks to complete
                    sleep(time: 60, unit: 'SECONDS')

                    // Check probe status
                    sh '''
                        export PATH="$PATH:$HOME/.local/bin"

                        PROBE_ID=$(stackeye probe list -o json | jq -r ".[] | select(.name | contains(\\"${APP_NAME}\\")) | .id" | head -1)

                        if [ -n "$PROBE_ID" ]; then
                            STATUS=$(stackeye probe get "$PROBE_ID" -o json | jq -r '.status')
                            echo "Probe status: $STATUS"

                            if [ "$STATUS" = "down" ]; then
                                echo "WARNING: Probe shows service is down!"
                                exit 1
                            fi
                        fi
                    '''
                }
            }
        }
    }

    post {
        failure {
            script {
                // Check for active alerts on failure
                sh '''
                    export PATH="$PATH:$HOME/.local/bin"

                    echo "Checking for active alerts..."
                    stackeye alert list --status active
                '''
            }
        }

        always {
            // Clean up
            cleanWs()
        }
    }
}

// Shared library function for reuse
// Add this to your Jenkins shared library

/*
def setupStackEyeMonitoring(Map config) {
    def appName = config.appName ?: 'my-app'
    def appUrl = config.appUrl ?: 'https://example.com'
    def interval = config.interval ?: 60
    def timeout = config.timeout ?: 10

    withCredentials([string(credentialsId: 'stackeye-api-key', variable: 'STACKEYE_API_KEY')]) {
        sh '''
            curl -fsSL https://get.stackeye.io/cli | bash
            export PATH="$PATH:$HOME/.local/bin"
            stackeye setup --no-input

            stackeye probe create \
                --name "''' + appName + ''' - Health" \
                --url "''' + appUrl + '''/health" \
                --interval ''' + interval + ''' \
                --timeout ''' + timeout + ''' || true
        '''
    }
}
*/
