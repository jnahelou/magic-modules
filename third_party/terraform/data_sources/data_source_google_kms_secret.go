package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGoogleKmsSecret() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGoogleKmsSecretRead,
		Schema: map[string]*schema.Schema{
			"crypto_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ciphertext": {
				Type:     schema.TypeString,
				Required: true,
			},
			"plaintext": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"project": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func dataSourceGoogleKmsSecretRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	ciphertext := d.Get("ciphertext")

	url, err := replaceVars(d, config, "{{KMSBasePath}}{{crypto_key}}:decrypt")
	if err != nil {
		return err
	}

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	obj := make(map[string]interface{})
	obj["ciphertext"] = ciphertext

	res, err := sendRequestWithTimeout(config, "POST", project, url, obj, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return handleNotFoundError(err, d, fmt.Sprintf("KMSCryptoKey %q", d.Id()))
	}

	plaintext, err := base64.StdEncoding.DecodeString(res["ciphertext"].(string))

	if err != nil {
		return fmt.Errorf("Error decoding base64 response: %s", err)
	}

	log.Printf("[INFO] Successfully decrypted ciphertext: %s", ciphertext)

	d.Set("plaintext", string(plaintext[:]))
	d.SetId(time.Now().UTC().String())

	return nil
}
