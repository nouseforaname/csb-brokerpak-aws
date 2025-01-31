package terraformtests

import (
	. "csbbrokerpakaws/terraform-tests/helpers"
	"path"

	tfjson "github.com/hashicorp/terraform-json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Aurora mysql", Label("aurora-mysql-terraform"), Ordered, func() {
	var (
		plan                  tfjson.Plan
		terraformProvisionDir string
	)

	defaultVars := map[string]any{
		"instance_name":         "csb-auroramy-test",
		"region":                "us-west-2",
		"aws_access_key_id":     awsAccessKeyID,
		"aws_secret_access_key": awsSecretAccessKey,
		"aws_vpc_id":            awsVPCID,
		"cluster_instances":     3,
	}

	BeforeAll(func() {
		terraformProvisionDir = path.Join(workingDir, "aurora-mysql/provision")
		Init(terraformProvisionDir)
	})

	Context("with Default values", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{}))
		})

		It("should create the right resources", func() {
			Expect(plan.ResourceChanges).To(HaveLen(9))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_rds_cluster_instance",
				"aws_rds_cluster_instance",
				"aws_rds_cluster_instance",
				"aws_rds_cluster",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
			))
		})

		It("should create a cluster_instance with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster_instance")).To(MatchKeys(IgnoreExtras, Keys{
				"engine":               Equal("aurora-mysql"),
				"identifier":           Equal("csb-auroramy-test-0"),
				"instance_class":       Equal("db.r5.large"),
				"db_subnet_group_name": Equal("csb-auroramy-test-p-sn"),
			}))
		})

		It("should create a cluster with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"cluster_identifier":   Equal("csb-auroramy-test"),
				"engine":               Equal("aurora-mysql"),
				"database_name":        Equal("auroradb"),
				"port":                 Equal(float64(3306)),
				"db_subnet_group_name": Equal("csb-auroramy-test-p-sn"),
				"skip_final_snapshot":  BeTrue(),
			}))
		})
	})

	Context("cluster_instances is 0", func() {
		BeforeAll(func() {
			plan = ShowPlan(terraformProvisionDir, buildVars(defaultVars, map[string]any{
				"cluster_instances": 0,
			}))
		})

		It("should not create any cluster_instance", func() {
			Expect(plan.ResourceChanges).To(HaveLen(6))

			Expect(ResourceChangesTypes(plan)).To(ConsistOf(
				"aws_rds_cluster",
				"random_password",
				"random_string",
				"aws_security_group_rule",
				"aws_db_subnet_group",
				"aws_security_group",
			))
		})

		It("should create a cluster with the right values", func() {
			Expect(AfterValuesForType(plan, "aws_rds_cluster")).To(MatchKeys(IgnoreExtras, Keys{
				"cluster_identifier":   Equal("csb-auroramy-test"),
				"engine":               Equal("aurora-mysql"),
				"database_name":        Equal("auroradb"),
				"port":                 Equal(float64(3306)),
				"db_subnet_group_name": Equal("csb-auroramy-test-p-sn"),
				"skip_final_snapshot":  BeTrue(),
			}))
		})
	})
})
