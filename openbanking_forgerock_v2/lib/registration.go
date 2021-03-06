package lib

// https://github.com/ramosbugs/openidconnect-rs/blob/master/src/registration.rs

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/banaio/openbankingforgerock/oidcdynamicclientregistration"
	"github.com/banaio/openbankingforgerock/requests"
)

const (
	// RegisterResponseFile the file to save the response from register.
	RegisterResponseFile = ".ignore/register_response.json"
)

// Register ..
func (c *OpenBankingClient) Register() (*oidcdynamicclientregistration.Response, error) {
	now := time.Now()
	iat := now.UTC().Unix()
	// exp := now.Add(24 * time.Hour).Unix()
	exp := time.Date(2022, 03, 29, 0, 0, 0, 0, time.UTC).Unix()
	// https://stackoverflow.com/questions/25065055/what-is-the-maximum-time-time-in-go
	// exp := time.Unix(1<<63-62135596801, 999999999).UTC().Unix()
	// exp := time.Unix(1<<60-62135596801, 999999999).UTC().Unix()
	jti := uuid.New().String()
	if exp != 0 {
	}

	claims := jwt.MapClaims{
		// "iss":                             c.Config.SSID,
		// "aud":                             c.OpenIDConfig.Issuer,
		// "aud":                             c.OpenIDConfig.TokenEndpoint,
		// "token_endpoint_auth_method":      "private_key_jwt",
		"exp":                             exp,
		"iss":                             c.Config.SSID,
		"software_statement":              c.Config.SSA,
		"iat":                             iat,
		"jti":                             jti,
		"token_endpoint_auth_signing_alg": "RS256",
		"request_object_encryption_alg":   "RSA-OAEP-256",
		"subject_type":                    "public",
		"application_type":                "web",
		"request_object_signing_alg":      "RS256",
		"request_object_encryption_enc":   "A128CBC-HS256",
		"id_token_signed_response_alg":    "RS256",
		"token_endpoint_auth_method":      "tls_client_auth",
		"scope":                           "openid accounts payments fundsconfirmations",
		"scopes": []string{
			"openid",
			"payments",
			"fundsconfirmations",
			"accounts",
		},
		"grant_types": []string{
			"refresh_token",
			"authorization_code",
			"client_credentials",
		},
		"redirect_uris": []string{
			"http://127.0.0.1:8080/openbanking/banaio",
			"http://127.0.0.1:8080/openbanking/banaio/forgerock",
			"http://127.0.0.1:8080/openbanking/banaio/callback",
			"http://127.0.0.1:8080/openbanking/banaio/forgerock/callback",
			"https://127.0.0.1:8080/openbanking/banaio",
			"https://127.0.0.1:8080/openbanking/banaio/forgerock",
			"https://127.0.0.1:8080/openbanking/banaio/callback",
			"https://127.0.0.1:8080/openbanking/banaio/forgerock/callback",

			"http://localhost:8080/openbanking/banaio",
			"http://localhost:8080/openbanking/banaio/forgerock",
			"http://localhost:8080/openbanking/banaio/callback",
			"http://localhost:8080/openbanking/banaio/forgerock/callback",
			"https://localhost:8080/openbanking/banaio",
			"https://localhost:8080/openbanking/banaio/forgerock",
			"https://localhost:8080/openbanking/banaio/callback",
			"https://localhost:8080/openbanking/banaio/forgerock/callback",

			"http://bana.io/openbanking/forgerock",
			"http://bana.io/openbanking/forgerock/callback",
			"https://bana.io/openbanking/forgerock",
			"https://bana.io/openbanking/forgerock/callback",

			"https://127.0.0.1:8443/conformancesuite/callback",
			"https://localhost:8443/conformancesuite/callback",
		},
		"response_types": []string{
			"code token id_token",
			"code",
			"code id_token",
			"device_code",
			"id_token",
			"code token",
			"token",
			"token id_token",
		},
	}
	jwt := c.Signer.Sign(claims)
	url := c.OpenIDConfig.RegistrationEndpoint

	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("Register")

	req, err := requests.NewRequest(http.MethodPost, url, bytes.NewBufferString(jwt))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("Register:NewRequest")
		return nil, err
	}
	req.Header.Set(requests.HeaderContentType, requests.MIMEApplicationJWT)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
			"jwt": jwt,
		}).Fatal("Register")
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 201 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"StatusCode": resp.StatusCode,
				"err":        err,
			}).Fatal("Register:ReadAll")
			return nil, err
		}
		defer resp.Body.Close()

		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"body":       string(body),
		}).Fatal("Register:ReadAll")
		return nil, err
	}

	if err := json.NewDecoder(resp.Body).Decode(c.RegisterResponse); err != nil {
		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"err":        err,
		}).Warn("Register:Decode")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"StatusCode": resp.StatusCode,
		"client_id":  c.RegisterResponse.ClientID,
	}).Info("Register")

	bytes, err := json.Marshal(c.RegisterResponse)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"err":        err,
		}).Fatal("Register:Marshal")
		return nil, err
	}

	if err := ioutil.WriteFile(RegisterResponseFile, bytes, 0644); err != nil {
		logrus.WithFields(logrus.Fields{
			"RegisterResponseFile": err,
			"bytes":                string(bytes),
			"err":                  err,
		}).Fatal("Register:ioutil.WriteFile")
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"RegisterResponseFile": RegisterResponseFile,
		"length":               len(bytes),
	}).Info("Register:WriteFile")

	return c.RegisterResponse, nil
}

// UnRegister ...
func (c *OpenBankingClient) UnRegister() {
	claims := jwt.MapClaims{
		"software_statement":         c.Config.SSA,
		"iss":                        c.Config.SSID,
		"aud":                        c.OpenIDConfig.Issuer,
		"token_endpoint_auth_method": "private_key_jwt",
		"request_object_signing_alg": "PS256",
		"grant_types": []string{
			"authorization_code",
			"refresh_token",
			"client_credentials",
		},
		"redirect_uris": []string{
			"http://127.0.0.1:8080/openbanking/banaio",
			"http://127.0.0.1:8080/openbanking/banaio/forgerock",
			"http://127.0.0.1:8080/openbanking/banaio/callback",
			"http://127.0.0.1:8080/openbanking/banaio/forgerock/callback",
			"https://127.0.0.1:8080/openbanking/banaio",
			"https://127.0.0.1:8080/openbanking/banaio/forgerock",
			"https://127.0.0.1:8080/openbanking/banaio/callback",
			"https://127.0.0.1:8080/openbanking/banaio/forgerock/callback",

			"http://localhost:8080/openbanking/banaio",
			"http://localhost:8080/openbanking/banaio/forgerock",
			"http://localhost:8080/openbanking/banaio/callback",
			"http://localhost:8080/openbanking/banaio/forgerock/callback",
			"https://localhost:8080/openbanking/banaio",
			"https://localhost:8080/openbanking/banaio/forgerock",
			"https://localhost:8080/openbanking/banaio/callback",
			"https://localhost:8080/openbanking/banaio/forgerock/callback",

			"http://bana.io/openbanking/forgerock",
			"http://bana.io/openbanking/forgerock/callback",
			"https://bana.io/openbanking/forgerock",
			"https://bana.io/openbanking/forgerock/callback",
		},
		"scopes": []string{
			"openid",
			"payments",
			"fundsconfirmations",
			"accounts",
		},
	}
	jwt := c.Signer.Sign(claims)
	// url := "https://rs.aspsp.ob.forgerock.financial/open-banking/registerTPP"
	url := c.OpenIDConfig.RegistrationEndpoint + "3c40836c-0c75-4044-b4fa-a5b17c00c91a"

	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("UnRegister")

	if jwt != "" {
	}
	// req, err := requests.NewRequest(http.MethodDelete, url, bytes.NewBufferString(jwt))
	req, err := requests.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Error("UnRegister:NewRequest")
		return
	}
	// req.Header.Set(requests.HeaderContentType, requests.MIMEApplicationJWT)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"err":        err,
		}).Error("UnRegister:Do")
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"err":        err,
		}).Error("UnRegister:ReadAll")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"StatusCode": resp.StatusCode,
			"Body":       string(body),
		}).Error("UnRegister")
		return
	}

	logrus.WithFields(logrus.Fields{
		"Body": string(body),
	}).Info("UnRegister")
}
