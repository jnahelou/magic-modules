package google

import (
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func dataSourceGoogleKmsSecretCiphertext() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGoogleKmsSecretCiphertextRead,
		Schema: map[string]*schema.Schema{
			"crypto_key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ciphertext": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"plaintext": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
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

func dataSourceGoogleKmsSecretCiphertextRead(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Config)

	project, err := getProject(d, config)
	if err != nil {
		return err
	}

	url, err := replaceVars(d, config, "{{KMSBasePath}}{{crypto_key}}:encrypt")
	if err != nil {
		return err
	}

	plaintext := base64.StdEncoding.EncodeToString([]byte(d.Get("plaintext").(string)))

	obj := make(map[string]interface{})
	obj["plaintext"] = plaintext

	res, err := sendRequestWithTimeout(config, "POST", project, url, obj, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return fmt.Errorf("Error encrypting plaintext: %s", err)
	}

	log.Printf("[INFO] Successfully encrypted plaintext")

	d.Set("name", res["name"].(string))
	d.Set("ciphertext", res["ciphertext"].(string))
	d.SetId(time.Now().UTC().String())

	return nil
}
