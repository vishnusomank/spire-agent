package helper

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/accuknox/auto-policy-discovery/src/cluster"
	logr "github.com/sirupsen/logrus"
	"github.com/spiffe/spire/pkg/agent"
	"github.com/spiffe/spire/pkg/common/catalog"
	"github.com/vishnusomank/spire-agent/internal/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CreateK8sSecrets(agentConf *agent.Config) error {

	path := getKeyPath(agentConf)

	client := cluster.ConnectK8sClient()

	keyFile, err := os.ReadFile(filepath.Clean(path))

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.SVID_SECRET_NAME,
		},
		Data: map[string][]byte{
			constants.KEYMANAGER_DISK_NAME: keyFile,
		},
		Type: v1.SecretTypeOpaque,
	}

	if err != nil {
		return err
	}

	_, err = client.CoreV1().Secrets(constants.SPIRE_NAMESPACE).Create(context.Background(), secret, metav1.CreateOptions{})

	if err != nil {
		return err
	}

	return nil
}

func getKeyPath(agentConf *agent.Config) string {
	pluginData, err := catalog.PluginConfigsFromHCL(agentConf.PluginConfigs)

	if err != nil {
		logr.Warn("Could not get plugin data")
		return ""
	}

	for _, data := range pluginData {
		if data.Name == constants.KEYMANAGER_DISK {
			path := strings.Split(data.Data, "=")[1]
			path = strings.TrimSpace(path)
			path = strings.TrimPrefix(path, "\"")
			path = strings.TrimSuffix(path, "\"")
			path = fmt.Sprintf("%s/%s", path, constants.KEYMANAGER_DISK_NAME)
			return path
		}
	}
	return ""
}

func WriteSVIDKey(agentConf *agent.Config) error {

	path := getKeyPath(agentConf)

	secret := GetK8sSecrets(agentConf)

	data := string(secret.Data[constants.KEYMANAGER_DISK_NAME])

	f, err := os.Create(filepath.Clean(path))
	if err != nil {
		logr.WithError(err).Error("File creation failed")
		return err
	}

	if _, err := f.WriteString(data); err != nil {
		logr.WithError(err).Error("WriteString failed")
	}
	if err := f.Sync(); err != nil {
		logr.WithError(err).Error("file sync failed")
	}
	if err := f.Close(); err != nil {
		logr.WithError(err).Error("file close failed")
	}

	logr.Infof("Found secret: %v in namespace: %v", secret.Name, secret.Namespace)
	logr.Infof("Writing secret to file: %v", path)
	return nil
}

func GetK8sSecrets(agentConf *agent.Config) v1.Secret {
	client := cluster.ConnectK8sClient()

	secrets, err := client.CoreV1().Secrets("").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		fmt.Printf("err: %v\n", err)
		return v1.Secret{}
	}

	for _, secret := range secrets.Items {

		if secret.Name == constants.SVID_SECRET_NAME {
			return secret
		}
	}
	return v1.Secret{}
}

func DeleteSVIDSecret() {
	client := cluster.ConnectK8sClient()

	err := client.CoreV1().Secrets(constants.SPIRE_NAMESPACE).Delete(context.Background(), constants.SVID_SECRET_NAME, metav1.DeleteOptions{})

	if err != nil {
		logr.WithError(err).Error("Failed to delete secret")
	}

}
