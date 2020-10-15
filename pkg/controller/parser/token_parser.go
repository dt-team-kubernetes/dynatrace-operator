package parser

import (
	"fmt"
	"strings"

	dynatracev1alpha1 "github.com/Dynatrace/dynatrace-operator/pkg/apis/dynatrace/v1alpha1"
	_const "github.com/Dynatrace/dynatrace-operator/pkg/controller/const"
	corev1 "k8s.io/api/core/v1"
)

type Tokens struct {
	ApiToken  string
	PaasToken string
}

func NewTokens(secret *corev1.Secret) (*Tokens, error) {
	if secret == nil {
		return nil, fmt.Errorf("could not parse tokens: secret is nil")
	}

	var apiToken string
	var paasToken string
	var err error

	if err = verifySecret(secret); err != nil {
		return nil, err
	}

	//Errors would have been caught by verifySecret
	apiToken, _ = ExtractToken(secret, _const.DynatraceApiToken)
	paasToken, _ = ExtractToken(secret, _const.DynatracePaasToken)

	return &Tokens{
		ApiToken:  apiToken,
		PaasToken: paasToken,
	}, nil
}

func verifySecret(secret *corev1.Secret) error {
	for _, token := range []string{_const.DynatracePaasToken, _const.DynatraceApiToken} {
		_, err := ExtractToken(secret, token)
		if err != nil {
			return fmt.Errorf("invalid secret %s, %s", secret.Name, err)
		}
	}

	return nil
}

func ExtractToken(secret *corev1.Secret, key string) (string, error) {
	value, ok := secret.Data[key]
	if !ok {
		err := fmt.Errorf("missing token %s", key)
		return "", err
	}

	return strings.TrimSpace(string(value)), nil
}

func GetTokensName(obj *dynatracev1alpha1.ActiveGate) string {
	if tkns := obj.Spec.Tokens; tkns != "" {
		return tkns
	}
	return obj.GetName()
}
