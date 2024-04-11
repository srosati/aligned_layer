package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/urfave/cli"

	"github.com/yetanotherco/aligned_layer/common"
	"github.com/yetanotherco/aligned_layer/core/config"
	"github.com/yetanotherco/aligned_layer/task_generator"
)

var (
	// Version is the version of the binary.
	Version   string
	GitCommit string
	GitDate   string
)

var (
	ConfigFileFlag = cli.StringFlag{
		Name:     "config",
		Required: false,
		Usage:    "Load configuration from `FILE`",
		Value:    "config-files/aggregator.yaml",
		EnvVar:   "CONFIG_FILE",
	}

	AlignedLayerDeploymentFileFlag = cli.StringFlag{
		Name:     "aligned-layer-deployment",
		Required: false,
		Usage:    "Load credible squaring contract addresses from `FILE`",
		Value:    "contracts/script/output/31337/aligned_layer_avs_deployment_output.json",
		EnvVar:   "ALIGNED_LAYER_DEPLOYMENT_FILE",
	}

	SharedAvsContractsDeploymentFileFlag = cli.StringFlag{
		Name:     "shared-avs-contracts-deployment",
		Required: false,
		Usage:    "Load shared avs contract addresses from `FILE`",
		Value:    "contracts/script/output/31337/shared_avs_contracts_deployment_output.json",
		EnvVar:   "SHARED_AVS_CONTRACTS_DEPLOYMENT_FILE",
	}

	EcdsaPrivateKeyFlag = cli.StringFlag{
		Name:     "ecdsa-private-key",
		Usage:    "Ethereum private key",
		Value:    "0x2a871d0798f97d79848a013d4936a73bf4cc922c825d33c1cf7073dff6d409c6",
		Required: false,
		EnvVar:   "ECDSA_PRIVATE_KEY",
	}

	ProofFileFlag = cli.StringFlag{
		Name:     "proof",
		Required: true,
		Usage:    "Load proof from `PROOF_FILE`",
	}

	VerifierIdFlag = cli.StringFlag{
		Name:     "verifier-id",
		Required: true,
		Usage:    "Set verifier ID",
	}

	PubInputIdFlag = cli.StringFlag{
		Name:     "pub-input",
		Required: false,
		Usage:    "Load public inputs from `PUB_INPUT_FILE`",
	}
)

var flags = []cli.Flag{
	ConfigFileFlag,
	AlignedLayerDeploymentFileFlag,
	SharedAvsContractsDeploymentFileFlag,
	EcdsaPrivateKeyFlag,
	ProofFileFlag,
	VerifierIdFlag,
	PubInputIdFlag,
}

func main() {
	app := cli.NewApp()
	app.Flags = flags
	app.Version = fmt.Sprintf("%s-%s-%s", Version, GitCommit, GitDate)
	app.Name = "Aligned Layer Task Sender"
	app.Usage = "Aligned Layer Task Sender"
	app.Description = "Service that sends proofs to verify by operator nodes."

	app.Action = taskSenderMain
	err := app.Run(os.Args)
	if err != nil {
		log.Fatalln("Task sender application failed.", "Message:", err)
	}
}

func taskSenderMain(ctx *cli.Context) error {
	log.Println("Initializing Task Sender...")

	log.Println("Config file: ", ctx.GlobalString(ConfigFileFlag.Name))
	config, err := config.NewConfig(ctx)
	if err != nil {
		return err
	}

	taskGen, err := task_generator.NewTaskGenerator(config)
	if err != nil {
		return err
	}

	proofFilePath := ctx.GlobalString(ProofFileFlag.Name)
	proof, err := os.ReadFile(proofFilePath)
	if err != nil {
		panic("Could not read proof file")
	}

	verifierId, err := parseVerifierId(ctx.GlobalString((VerifierIdFlag.Name)))
	if err != nil {
		return err
	}

	var pubInput []byte
	// When we have a PLONK or Kimchi proof, we should check for the public inputs.
	// Cairo proofs have public inputs embedded, so no need to check for this CLI input for the moment.
	// This should be done for every proving system.
	if verifierId == common.GnarkPlonkBls12_381 || verifierId == common.Kimchi {
		pubInputFilePath := ctx.GlobalString(PubInputIdFlag.Name)
		pubInput, err = os.ReadFile(pubInputFilePath)
		if err != nil {
			panic("Could not public input file")
		}
	}

	err = taskGen.SendNewTask(proof, pubInput, verifierId)
	if err != nil {
		return err
	}

	log.Println("Task successfully sent")

	return nil
}

func parseVerifierId(verifierIdStr string) (common.VerifierId, error) {
	// standard whitespace trimming
	verifierIdStr = strings.TrimSpace(verifierIdStr)
	switch verifierIdStr {
	case "cairo":
		return common.LambdaworksCairo, nil
	case "plonk":
		return common.GnarkPlonkBls12_381, nil
	case "kimchi":
		return common.Kimchi, nil
	case "sp1":
		return common.Sp1BabyBearBlake3, nil
	default:
		// returning this just to return something, the error should be handled
		// by the caller.
		return common.LambdaworksCairo, errors.New("could not parse verifier ID")
	}
}
