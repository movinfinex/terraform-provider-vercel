package vercel_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/vercel/terraform-provider-vercel/client"
)

func TestAcc_Project(t *testing.T) {
	testTeamID := resource.TestCheckNoResourceAttr("vercel_project.test", "team_id")
	if testTeam() != "" {
		testTeamID = resource.TestCheckResourceAttr("vercel_project.test", "team_id", testTeam())
	}
	projectSuffix := acctest.RandString(16)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("vercel_project.test", testTeam()),
		Steps: []resource.TestStep{
			// Ensure we get nice framework / serverless_function_region errors
			{
				Config: `
                    resource "vercel_project" "test" {
                        name = "foo"
                        serverless_function_region = "notexist"
                    }
                `,
				ExpectError: regexp.MustCompile("Invalid Serverless Function Region"),
			},
			{
				Config: `
                    resource "vercel_project" "test" {
                        name = "foo"
                        framework = "notexist"
                    }
                `,
				ExpectError: regexp.MustCompile("Invalid Framework"),
			},
			// Create and Read testing
			{
				Config: testAccProjectConfig(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test", testTeam()),
					testTeamID,
					resource.TestCheckResourceAttr("vercel_project.test", "name", fmt.Sprintf("test-acc-project-%s", projectSuffix)),
					resource.TestCheckResourceAttr("vercel_project.test", "build_command", "npm run build"),
					resource.TestCheckResourceAttr("vercel_project.test", "dev_command", "npm run serve"),
					resource.TestCheckResourceAttr("vercel_project.test", "framework", "nextjs"),
					resource.TestCheckResourceAttr("vercel_project.test", "install_command", "npm install"),
					resource.TestCheckResourceAttr("vercel_project.test", "output_directory", ".output"),
					resource.TestCheckResourceAttr("vercel_project.test", "public_source", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "root_directory", "ui/src"),
					resource.TestCheckResourceAttr("vercel_project.test", "ignore_command", "echo 'wat'"),
					resource.TestCheckResourceAttr("vercel_project.test", "serverless_function_region", "syd1"),
					resource.TestCheckResourceAttr("vercel_project.test", "automatically_expose_system_environment_variables", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.test", "environment.*", map[string]string{
						"key":   "foo",
						"value": "bar",
					}),
					resource.TestCheckTypeSetElemAttr("vercel_project.test", "environment.0.target.*", "production"),
					resource.TestCheckResourceAttr("vercel_project.test", "git_comments.on_pull_request", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "git_comments.on_commit", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "preview_comments", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "auto_assign_custom_domains", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "git_lfs", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "function_failover", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "customer_success_code_visibility", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "git_fork_protection", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "prioritise_production_builds", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "directory_listing", "true"),
					resource.TestCheckResourceAttr("vercel_project.test", "skew_protection", "7 days"),
				),
			},
			// Update testing
			{
				Config: testAccProjectConfigUpdated(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vercel_project.test", "name", fmt.Sprintf("test-acc-two-%s", projectSuffix)),
					resource.TestCheckNoResourceAttr("vercel_project.test", "build_command"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.test", "environment.*", map[string]string{
						"key":   "bar",
						"value": "baz",
					}),
				),
			},
		},
	})
}

func TestAcc_ProjectAddingEnvAfterInitialCreation(t *testing.T) {
	projectSuffix := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("vercel_project.test", testTeam()),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfigWithoutEnv(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test", testTeam()),
				),
			},
			{
				Config: testAccProjectConfigUpdated(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test", testTeam()),
				),
			},
		},
	})
}

func TestAcc_ProjectWithGitRepository(t *testing.T) {
	projectSuffix := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("vercel_project.test_git", testTeam()),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfigWithGitRepo(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test_git", testTeam()),
					resource.TestCheckResourceAttr("vercel_project.test_git", "git_repository.type", "github"),
					resource.TestCheckResourceAttr("vercel_project.test_git", "git_repository.repo", testGithubRepo()),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.test_git", "environment.*", map[string]string{
						"key":        "foo",
						"value":      "bar",
						"git_branch": "staging",
					}),
				),
			},
			{
				Config: testAccProjectConfigWithGitRepoUpdated(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test_git", testTeam()),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.test_git", "environment.*", map[string]string{
						"key":   "foo",
						"value": "bar2",
					}),
				),
			},
			{
				Config: testAccProjectConfigWithGitRepoRemoved(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test_git", testTeam()),
					resource.TestCheckNoResourceAttr("vercel_project.test_git", "git_repository"),
				),
			},
		},
	})
}

func TestAcc_ProjectWithVercelAuthAndPasswordProtectionAndTrustedIps(t *testing.T) {
	projectSuffix := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("vercel_project.enabled_to_start", testTeam()),
		Steps: []resource.TestStep{
			{
				Config: testAccProjectConfigWithVercelAuthAndPasswordAndTrustedIps(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.enabled_to_start", testTeam()),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "vercel_authentication.deployment_type", "all_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "password_protection.deployment_type", "all_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "password_protection.password", "password"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "trusted_ips.addresses.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.enabled_to_start", "trusted_ips.addresses.*", map[string]string{
						"value": "1.1.1.1",
						"note":  "notey note",
					}),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "trusted_ips.deployment_type", "all_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "trusted_ips.protection_mode", "trusted_ip_optional"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "protection_bypass_for_automation", "true"),
					resource.TestCheckResourceAttrSet("vercel_project.enabled_to_start", "protection_bypass_for_automation_secret"),
					testAccProjectExists("vercel_project.disabled_to_start", testTeam()),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "vercel_authentication.deployment_type", "standard_protection"),
					resource.TestCheckNoResourceAttr("vercel_project.disabled_to_start", "password_protection"),
					resource.TestCheckNoResourceAttr("vercel_project.disabled_to_start", "trusted_ips"),
					resource.TestCheckNoResourceAttr("vercel_project.disabled_to_start", "protection_bypass_for_automation"),
					resource.TestCheckNoResourceAttr("vercel_project.disabled_to_start", "protection_bypass_for_automation_secret"),
					testAccProjectExists("vercel_project.enabled_to_update", testTeam()),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "vercel_authentication.deployment_type", "only_preview_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "password_protection.deployment_type", "only_preview_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "password_protection.password", "password"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.addresses.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.enabled_to_update", "trusted_ips.addresses.*", map[string]string{
						"value": "1.1.1.3",
						"note":  "notey notey note",
					}),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.deployment_type", "only_production_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.protection_mode", "trusted_ip_required"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "protection_bypass_for_automation", "true"),
					resource.TestCheckResourceAttrSet("vercel_project.enabled_to_update", "protection_bypass_for_automation_secret"),
				),
			},
			{
				Config: testAccProjectConfigWithVercelAuthAndPasswordAndTrustedIpsUpdated(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vercel_project.enabled_to_start", "vercel_authentication.deployment_type", "standard_protection"),
					resource.TestCheckNoResourceAttr("vercel_project.enabled_to_start", "password_protection"),
					resource.TestCheckNoResourceAttr("vercel_project.enabled_to_start", "protection_bypass_for_automation"),
					resource.TestCheckNoResourceAttr("vercel_project.enabled_to_start", "trusted_ips"),
					resource.TestCheckNoResourceAttr("vercel_project.enabled_to_start", "protection_bypass_for_automation_secret"),

					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "vercel_authentication.deployment_type", "standard_protection"),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "password_protection.deployment_type", "standard_protection"),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "password_protection.password", "password"),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "trusted_ips.addresses.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.disabled_to_start", "trusted_ips.addresses.*", map[string]string{
						"value": "1.1.1.1",
						"note":  "notey note",
					}),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "trusted_ips.deployment_type", "standard_protection"),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "trusted_ips.protection_mode", "trusted_ip_required"),
					resource.TestCheckResourceAttr("vercel_project.disabled_to_start", "protection_bypass_for_automation", "true"),
					resource.TestCheckResourceAttrSet("vercel_project.disabled_to_start", "protection_bypass_for_automation_secret"),

					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "vercel_authentication.deployment_type", "standard_protection"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "password_protection.deployment_type", "standard_protection"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "password_protection.password", "password2"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.addresses.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vercel_project.enabled_to_update", "trusted_ips.addresses.*", map[string]string{
						"value": "1.1.1.3",
						"note":  "notey notey",
					}),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.deployment_type", "all_deployments"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "trusted_ips.protection_mode", "trusted_ip_optional"),
					resource.TestCheckResourceAttr("vercel_project.enabled_to_update", "protection_bypass_for_automation", "false"),
					resource.TestCheckNoResourceAttr("vercel_project.enabled_to_update", "protection_bypass_for_automation_secret"),
				),
			},
		},
	})
}

func getProjectImportID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set")
		}

		if rs.Primary.Attributes["team_id"] == "" {
			return rs.Primary.ID, nil
		}
		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["team_id"], rs.Primary.ID), nil
	}
}

func TestAcc_ProjectImport(t *testing.T) {
	projectSuffix := acctest.RandString(16)
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccProjectDestroy("vercel_project.test", testTeam()),
		Steps: []resource.TestStep{
			{
				Config: projectConfigWithoutEnv(projectSuffix, teamIDConfig()),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccProjectExists("vercel_project.test", testTeam()),
				),
			},
			{
				ResourceName:      "vercel_project.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: getProjectImportID("vercel_project.test"),
			},
		},
	})
}

func testAccProjectExists(n, teamID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no projectID is set")
		}

		_, err := testClient().GetProject(context.TODO(), rs.Primary.ID, teamID)
		return err
	}
}

func testAccProjectDestroy(n, teamID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no projectID is set")
		}

		_, err := testClient().GetProject(context.TODO(), rs.Primary.ID, teamID)
		if err == nil {
			return fmt.Errorf("expected not_found error, but got no error")
		}
		if !client.NotFound(err) {
			return fmt.Errorf("Unexpected error checking for deleted project: %s", err)
		}

		return nil
	}
}

func testAccProjectConfigWithoutEnv(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test" {
  name = "test-acc-two-%s"
  %s
}
`, projectSuffix, teamID)
}

func testAccProjectConfigUpdated(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test" {
  name = "test-acc-two-%s"
  %s
  environment = [
    {
      key    = "two"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "foo"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "baz"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "three"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "oh_no"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "bar"
      value  = "baz"
      target = ["production"]
    },
    {
      key       = "sensitive_thing"
      value     = "bar_updated"
      target    = ["production"]
      sensitive = true
    }
  ]
}
`, projectSuffix, teamID)
}

func testAccProjectConfigWithVercelAuthAndPasswordAndTrustedIps(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "enabled_to_start" {
  name = "test-acc-protection-one-%[1]s"
  %[2]s
  vercel_authentication = {
    deployment_type = "all_deployments"
  }
  password_protection = {
    deployment_type = "all_deployments"
    password           = "password"
  }
  trusted_ips = {
	addresses = [
		{
			value = "1.1.1.1"
			note = "notey note"
		}
	]
	deployment_type = "all_deployments"
	protection_mode = "trusted_ip_optional"
  }
  protection_bypass_for_automation = true
}

resource "vercel_project" "disabled_to_start" {
  name = "test-acc-protection-two-%[1]s"
  %[2]s
}

resource "vercel_project" "enabled_to_update" {
  name = "test-acc-protection-three-%[1]s"
  %[2]s
  vercel_authentication = {
    deployment_type = "only_preview_deployments"
  }
  password_protection = {
    deployment_type = "only_preview_deployments"
    password           = "password"
  }
  trusted_ips = {
	addresses = [
		{
			value = "1.1.1.1"
			note = "notey notey"
		},
		{
			value = "1.1.1.3"
			note = "notey notey note"
		}
	]
	deployment_type = "only_production_deployments"
  }
  protection_bypass_for_automation = true
}
    `, projectSuffix, teamID)
}

func testAccProjectConfigWithVercelAuthAndPasswordAndTrustedIpsUpdated(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "enabled_to_start" {
  name = "test-acc-protection-one-%[1]s"
  %[2]s
}

resource "vercel_project" "disabled_to_start" {
  name = "test-acc-protection-two-%[1]s"
  %[2]s
  vercel_authentication = {
    deployment_type = "standard_protection"
  }
  password_protection = {
    deployment_type = "standard_protection"
    password           = "password"
  }
  trusted_ips = {
	addresses = [
		{
			value = "1.1.1.1"
			note = "notey note"
		}
	]
	deployment_type = "standard_protection"
  }
  protection_bypass_for_automation = true
}

resource "vercel_project" "enabled_to_update" {
  name = "test-acc-protection-three-%[1]s"
  %[2]s
  vercel_authentication = {
    deployment_type = "standard_protection"
  }
  password_protection = {
    deployment_type = "standard_protection"
    password           = "password2"
  }
  trusted_ips = {
	addresses = [
		{
			value = "1.1.1.3"
			note = "notey notey"
		}
	]
	deployment_type = "all_deployments"
	protection_mode = "trusted_ip_optional"
  }
  protection_bypass_for_automation = false
}
    `, projectSuffix, teamID)
}

func testAccProjectConfigWithGitRepo(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test_git" {
  name = "test-acc-two-%s"
  %s
  git_repository = {
    type = "github"
    repo = "%s"
    deploy_hooks = [
        {
            ref = "main"
            name = "some deploy hook"
        }
    ]
  }
  environment = [
    {
      key        = "foo"
      value      = "bar"
      target     = ["preview"]
      git_branch = "staging"
    }
  ]
}
    `, projectSuffix, teamID, testGithubRepo())
}

func testAccProjectConfigWithGitRepoUpdated(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test_git" {
  name = "test-acc-two-%s"
  %s
  public_source = false
  git_repository = {
    type = "github"
    repo = "%s"
    production_branch = "staging"
    deploy_hooks = [
        {
            ref = "main"
            name = "some deploy hook"
        },
        {
            ref = "main"
            name = "some other hook"
        }
    ]
  }
  environment = [
    {
      key        = "foo"
      value      = "bar2"
      target     = ["preview"]
    }
  ]
}
    `, projectSuffix, teamID, testGithubRepo())
}

func testAccProjectConfigWithGitRepoRemoved(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test_git" {
  name = "test-acc-two-%s"
  %s
  public_source = false
  environment = [
    {
      key        = "foo"
      value      = "bar2"
      target     = ["preview"]
    }
  ]
}
    `, projectSuffix, teamID)
}

func projectConfigWithoutEnv(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test" {
  name = "test-acc-project-%s"
  %s
  build_command = "npm run build"
  dev_command = "npm run serve"
  ignore_command = "echo 'wat'"
  serverless_function_region = "syd1"
  framework = "nextjs"
  install_command = "npm install"
  output_directory = ".output"
  public_source = true
  root_directory = "ui/src"
}
`, projectSuffix, teamID)
}

func testAccProjectConfig(projectSuffix, teamID string) string {
	return fmt.Sprintf(`
resource "vercel_project" "test" {
  name = "test-acc-project-%s"
  %s
  build_command = "npm run build"
  dev_command = "npm run serve"
  ignore_command = "echo 'wat'"
  serverless_function_region = "syd1"
  framework = "nextjs"
  install_command = "npm install"
  output_directory = ".output"
  public_source = true
  root_directory = "ui/src"
  automatically_expose_system_environment_variables = true
  git_comments = {
      on_pull_request = true,
      on_commit = true
  }
  preview_comments = true
  auto_assign_custom_domains = true
  git_lfs = true
  function_failover = true
  customer_success_code_visibility = true
  git_fork_protection = true
  prioritise_production_builds = true
  directory_listing = true
  skew_protection = "7 days"

  environment = [
    {
      key    = "foo"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "two"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "three"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "baz"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "bar"
      value  = "bar"
      target = ["production"]
    },
    {
      key    = "oh_no"
      value  = "bar"
      target = ["production"]
    },
    {
      key       = "sensitive_thing"
      value     = "bar"
      target    = ["production"]
      sensitive = true
    }
  ]
}
`, projectSuffix, teamID)
}
