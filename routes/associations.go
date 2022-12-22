package routes

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/deso-protocol/core/lib"
	"github.com/gorilla/mux"
	"io"
	"net/http"
)

// ------------
// Types
// ------------

type CreateUserAssociationRequest struct {
	TransactorPublicKeyBase58Check string            `safeForLogging:"true"`
	TargetUserPublicKeyBase58Check string            `safeForLogging:"true"`
	AppPublicKeyBase58Check        string            `safeForLogging:"true"`
	AssociationType                string            `safeForLogging:"true"`
	AssociationValue               string            `safeForLogging:"true"`
	ExtraData                      map[string]string `safeForLogging:"true"`
	MinFeeRateNanosPerKB           uint64            `safeForLogging:"true"`
	TransactionFees                []TransactionFee  `safeForLogging:"true"`
}

type UserAssociationQuery struct {
	TransactorPublicKeyBase58Check string `safeForLogging:"true"`
	TargetUserPublicKeyBase58Check string `safeForLogging:"true"`
	AppPublicKeyBase58Check        string `safeForLogging:"true"`
	AssociationType                string `safeForLogging:"true"`
	AssociationTypePrefix          string `safeForLogging:"true"`
	AssociationValue               string `safeForLogging:"true"`
	AssociationValuePrefix         string `safeForLogging:"true"`
	Limit                          int    `safeForLogging:"true"`
	LastSeenAssociationID          string `safeForLogging:"true"`
	SortDescending                 bool   `safeForLogging:"true"`
}

type UserAssociationResponse struct {
	AssociationID                  string            `safeForLogging:"true"`
	TransactorPublicKeyBase58Check string            `safeForLogging:"true"`
	TargetUserPublicKeyBase58Check string            `safeForLogging:"true"`
	AppPublicKeyBase58Check        string            `safeForLogging:"true"`
	AssociationType                string            `safeForLogging:"true"`
	AssociationValue               string            `safeForLogging:"true"`
	ExtraData                      map[string]string `safeForLogging:"true"`
	BlockHeight                    uint32            `safeForLogging:"true"`
}

type UserAssociationsResponse struct {
	Associations []*UserAssociationResponse
}

type CreatePostAssociationRequest struct {
	TransactorPublicKeyBase58Check string            `safeForLogging:"true"`
	PostHashHex                    string            `safeForLogging:"true"`
	AppPublicKeyBase58Check        string            `safeForLogging:"true"`
	AssociationType                string            `safeForLogging:"true"`
	AssociationValue               string            `safeForLogging:"true"`
	ExtraData                      map[string]string `safeForLogging:"true"`
	MinFeeRateNanosPerKB           uint64            `safeForLogging:"true"`
	TransactionFees                []TransactionFee  `safeForLogging:"true"`
}

type PostAssociationQuery struct {
	TransactorPublicKeyBase58Check string `safeForLogging:"true"`
	PostHashHex                    string `safeForLogging:"true"`
	AppPublicKeyBase58Check        string `safeForLogging:"true"`
	AssociationType                string `safeForLogging:"true"`
	AssociationTypePrefix          string `safeForLogging:"true"`
	AssociationValue               string `safeForLogging:"true"`
	AssociationValuePrefix         string `safeForLogging:"true"`
	Limit                          int    `safeForLogging:"true"`
	LastSeenAssociationID          string `safeForLogging:"true"`
	SortDescending                 bool   `safeForLogging:"true"`
}

type PostAssociationResponse struct {
	AssociationID                  string            `safeForLogging:"true"`
	TransactorPublicKeyBase58Check string            `safeForLogging:"true"`
	PostHashHex                    string            `safeForLogging:"true"`
	AppPublicKeyBase58Check        string            `safeForLogging:"true"`
	AssociationType                string            `safeForLogging:"true"`
	AssociationValue               string            `safeForLogging:"true"`
	ExtraData                      map[string]string `safeForLogging:"true"`
	BlockHeight                    uint32            `safeForLogging:"true"`
}

type PostAssociationsResponse struct {
	Associations []*PostAssociationResponse
}

type DeleteAssociationRequest struct {
	TransactorPublicKeyBase58Check string            `safeForLogging:"true"`
	AssociationID                  string            `safeForLogging:"true"`
	ExtraData                      map[string]string `safeForLogging:"true"`
	MinFeeRateNanosPerKB           uint64            `safeForLogging:"true"`
	TransactionFees                []TransactionFee  `safeForLogging:"true"`
}

type AssociationTxnResponse struct {
	SpendAmountNanos  uint64
	TotalInputNanos   uint64
	ChangeAmountNanos uint64
	FeeNanos          uint64
	Transaction       *lib.MsgDeSoTxn
	TransactionHex    string
	TxnHashHex        string
}

// ------------
// Routes
// ------------

func (fes *APIServer) CreateUserAssociation(ww http.ResponseWriter, req *http.Request) {
	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := CreateUserAssociationRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "CreateUserAssociation: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem getting UTXO view")
		return
	}

	// Parse TransactorPublicKeyBytes from TransactorPublicKeyBase58Check.
	if requestData.TransactorPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "CreateUserAssociation: must provide a TransactorPublicKeyBase58Check")
		return
	}
	transactorPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.TransactorPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem getting public key for the transactor")
		return
	}

	// Parse TargetUserPublicKeyBytes from TargetUserPublicKeyBase58Check.
	if requestData.TargetUserPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "CreateUserAssociation: must provide a TargetUserPublicKeyBase58Check")
		return
	}
	targetUserPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.TargetUserPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem getting public key for the target user")
		return
	}

	// Parse AppPublicKeyBytes from AppPublicKeyBase58Check
	if requestData.AppPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "CreateUserAssociation: must provide an AppPublicKeyBase58Check")
		return
	}
	appPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.AppPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem getting public key for the app")
		return
	}

	// Validate AssociationType.
	if requestData.AssociationType == "" {
		_AddBadRequestError(ww, "CreateUserAssociation: must provide an AssociationType")
		return
	}

	// Validate AssociationValue.
	if requestData.AssociationValue == "" {
		_AddBadRequestError(ww, "CreateUserAssociation: must provide an AssociationValue")
		return
	}

	// Parse ExtraData.
	extraData, err := EncodeExtraDataMap(requestData.ExtraData)
	if err != nil {
		_AddBadRequestError(ww, "CreateUserAssociation: invalid ExtraData provided")
		return
	}

	// Compute the additional transaction fees as specified
	// by the request body and the node-level fees.
	additionalOutputs, err := fes.getTransactionFee(
		lib.TxnTypeCreateUserAssociation,
		transactorPublicKeyBytes,
		requestData.TransactionFees,
	)
	if err != nil {
		_AddBadRequestError(ww, "CreateUserAssociation: specified TransactionFees are invalid")
		return
	}

	// Create transaction.
	txn, totalInput, changeAmount, fees, err := fes.blockchain.CreateCreateUserAssociationTxn(
		transactorPublicKeyBytes,
		&lib.CreateUserAssociationMetadata{
			TargetUserPublicKey: lib.NewPublicKey(targetUserPublicKeyBytes),
			AppPublicKey:        lib.NewPublicKey(appPublicKeyBytes),
			AssociationType:     []byte(requestData.AssociationType),
			AssociationValue:    []byte(requestData.AssociationValue),
		},
		extraData,
		requestData.MinFeeRateNanosPerKB,
		fes.backendServer.GetMempool(),
		additionalOutputs,
	)
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("CreateUserAssociation: problem creating txn: %v", err))
		return
	}

	// Construct response.
	txnBytes, err := txn.ToBytes(true)
	if err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem encoding txn to bytes")
		return
	}
	res := AssociationTxnResponse{
		SpendAmountNanos:  totalInput - changeAmount - fees,
		TotalInputNanos:   totalInput,
		ChangeAmountNanos: changeAmount,
		FeeNanos:          fees,
		Transaction:       txn,
		TransactionHex:    hex.EncodeToString(txnBytes),
		TxnHashHex:        txn.Hash().String(),
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, "CreateUserAssociation: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) DeleteUserAssociation(ww http.ResponseWriter, req *http.Request) {
	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := DeleteAssociationRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "DeleteUserAssociation: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "DeleteUserAssociation: problem getting UTXO view")
		return
	}

	// Parse TransactorPublicKeyBytes from TransactorPublicKeyBase58Check.
	if requestData.TransactorPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "DeleteUserAssociation: must provide a TransactorPublicKeyBase58Check")
		return
	}
	transactorPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.TransactorPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "DeleteUserAssociation: problem getting public key for the transactor")
		return
	}

	// Parse AssociationIDBytes from AssociationID (hex string).
	if requestData.AssociationID == "" {
		_AddBadRequestError(ww, "DeleteUserAssociation: must provide an AssociationID")
		return
	}
	associationIdBytes, err := hex.DecodeString(requestData.AssociationID)
	if err != nil {
		_AddBadRequestError(ww, "DeleteUserAssociation: invalid AssociationID provided")
		return
	}

	// Parse ExtraData.
	extraData, err := EncodeExtraDataMap(requestData.ExtraData)
	if err != nil {
		_AddBadRequestError(ww, "DeleteUserAssociation: invalid ExtraData provided")
		return
	}

	// Compute the additional transaction fees as specified
	// by the request body and the node-level fees.
	additionalOutputs, err := fes.getTransactionFee(
		lib.TxnTypeDeleteUserAssociation,
		transactorPublicKeyBytes,
		requestData.TransactionFees,
	)
	if err != nil {
		_AddBadRequestError(ww, "DeleteUserAssociation: specified TransactionFees are invalid")
		return
	}

	// Create transaction.
	txn, totalInput, changeAmount, fees, err := fes.blockchain.CreateDeleteUserAssociationTxn(
		transactorPublicKeyBytes,
		&lib.DeleteUserAssociationMetadata{AssociationID: lib.NewBlockHash(associationIdBytes)},
		extraData,
		requestData.MinFeeRateNanosPerKB,
		fes.backendServer.GetMempool(),
		additionalOutputs,
	)
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("DeleteUserAssociation: problem creating txn: %v", err))
		return
	}

	// Construct response.
	txnBytes, err := txn.ToBytes(true)
	if err != nil {
		_AddInternalServerError(ww, "DeleteUserAssociation: problem encoding txn to bytes")
		return
	}
	res := AssociationTxnResponse{
		SpendAmountNanos:  totalInput - changeAmount - fees,
		TotalInputNanos:   totalInput,
		ChangeAmountNanos: changeAmount,
		FeeNanos:          fees,
		Transaction:       txn,
		TransactionHex:    hex.EncodeToString(txnBytes),
		TxnHashHex:        txn.Hash().String(),
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, "DeleteUserAssociation: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) GetUserAssociationByID(ww http.ResponseWriter, req *http.Request) {
	// Parse AssociationID from URL.
	vars := mux.Vars(req)
	associationIdHex, associationIdExists := vars["associationID"]
	if !associationIdExists {
		_AddBadRequestError(ww, fmt.Sprintf("GetUserAssociationByID: must provide an AssociationID"))
		return
	}

	// Parse AssociationID (BlockHash) from AssociationIdHex (string).
	associationIdBytes, err := hex.DecodeString(associationIdHex)
	if err != nil {
		_AddBadRequestError(ww, "GetUserAssociationByID: invalid AssociationID provided")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "GetUserAssociationByID: problem getting UTXO view")
		return
	}

	// Fetch AssociationEntry.
	associationEntry, err := utxoView.GetUserAssociationByID(lib.NewBlockHash(associationIdBytes))
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetUserAssociationByID: problem retrieving association: %v", err))
		return
	}
	if associationEntry == nil {
		_AddBadRequestError(ww, "GetUserAssociationByID: association not found")
		return
	}

	// Convert AssociationEntry to AssociationResponse.
	response := UserAssociationResponse{
		AssociationID:                  hex.EncodeToString(associationEntry.AssociationID.ToBytes()),
		TransactorPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TransactorPKID), false, fes.Params),
		TargetUserPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TargetUserPKID), false, fes.Params),
		AppPublicKeyBase58Check:        lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.AppPKID), false, fes.Params),
		AssociationType:                string(associationEntry.AssociationType),
		AssociationValue:               string(associationEntry.AssociationValue),
		ExtraData:                      DecodeExtraDataMap(fes.Params, utxoView, associationEntry.ExtraData),
		BlockHeight:                    associationEntry.BlockHeight,
	}

	// JSON encode response.
	if err = json.NewEncoder(ww).Encode(response); err != nil {
		_AddInternalServerError(ww, "GetUserAssociationByID: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) GetUserAssociations(ww http.ResponseWriter, req *http.Request) {
	var err error

	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := UserAssociationQuery{}
	if err = decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "GetUserAssociations: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "GetUserAssociations: problem getting UTXO view")
		return
	}

	// Parse TransactorPKID from TransactorPublicKeyBase58Check.
	var transactorPKID *lib.PKID
	if requestData.TransactorPublicKeyBase58Check != "" {
		transactorPKID, err = fes.getPKIDFromPublicKeyBase58Check(
			utxoView, requestData.TransactorPublicKeyBase58Check,
		)
		if err != nil {
			_AddInternalServerError(ww, "GetUserAssociations: problem getting PKID for the transactor")
			return
		}
	}

	// Parse TargetUserPKID from TargetUserPublicKeyBase58Check.
	var targetUserPKID *lib.PKID
	if requestData.TargetUserPublicKeyBase58Check != "" {
		targetUserPKID, err = fes.getPKIDFromPublicKeyBase58Check(
			utxoView, requestData.TargetUserPublicKeyBase58Check,
		)
		if err != nil {
			_AddInternalServerError(ww, "GetUserAssociations: problem getting PKID for the target user")
			return
		}
	}

	// Parse AppPKID from AppPublicKeyBase58Check
	var appPKID *lib.PKID
	if requestData.AppPublicKeyBase58Check != "" {
		appPKID, err = fes.getPKIDFromPublicKeyBase58Check(
			utxoView, requestData.AppPublicKeyBase58Check,
		)
		if err != nil {
			_AddInternalServerError(ww, "GetUserAssociations: problem getting PKID for the app")
			return
		}
	}

	// Parse LastSeenAssociationID (BlockHash) from LastSeenAssociationIdHex (string).
	var lastSeenAssociationID *lib.BlockHash
	if requestData.LastSeenAssociationID != "" {
		lastSeenAssociationIdBytes, err := hex.DecodeString(requestData.LastSeenAssociationID)
		if err != nil {
			_AddBadRequestError(ww, "GetUserAssociations: invalid LastSeenAssociationID provided")
			return
		}
		lastSeenAssociationID = lib.NewBlockHash(lastSeenAssociationIdBytes)
	}

	// Query for association entries.
	associationEntries, err := utxoView.GetUserAssociationsByAttributes(&lib.UserAssociationQuery{
		TransactorPKID:         transactorPKID,
		TargetUserPKID:         targetUserPKID,
		AppPKID:                appPKID,
		AssociationType:        []byte(requestData.AssociationType),
		AssociationTypePrefix:  []byte(requestData.AssociationTypePrefix),
		AssociationValue:       []byte(requestData.AssociationValue),
		AssociationValuePrefix: []byte(requestData.AssociationValuePrefix),
		Limit:                  requestData.Limit,
		LastSeenAssociationID:  lastSeenAssociationID,
		SortDescending:         requestData.SortDescending,
	})
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetUserAssociations: %v", err))
		return
	}

	// Convert AssociationEntries to AssociationResponses.
	associationResponses := []*UserAssociationResponse{}
	for _, associationEntry := range associationEntries {
		associationResponses = append(associationResponses, &UserAssociationResponse{
			AssociationID:                  hex.EncodeToString(associationEntry.AssociationID.ToBytes()),
			TransactorPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TransactorPKID), false, fes.Params),
			TargetUserPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TargetUserPKID), false, fes.Params),
			AppPublicKeyBase58Check:        lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.AppPKID), false, fes.Params),
			AssociationType:                string(associationEntry.AssociationType),
			AssociationValue:               string(associationEntry.AssociationValue),
			ExtraData:                      DecodeExtraDataMap(fes.Params, utxoView, associationEntry.ExtraData),
			BlockHeight:                    associationEntry.BlockHeight,
		})
	}

	// JSON encode response.
	response := UserAssociationsResponse{Associations: associationResponses}
	if err = json.NewEncoder(ww).Encode(response); err != nil {
		_AddInternalServerError(ww, "GetUserAssociations: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) CreatePostAssociation(ww http.ResponseWriter, req *http.Request) {
	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := CreatePostAssociationRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "CreatePostAssociation: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "CreatePostAssociation: problem getting UTXO view")
		return
	}

	// Parse TransactorPublicKeyBytes from TransactorPublicKeyBase58Check.
	if requestData.TransactorPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "CreatePostAssociation: must provide a TransactorPublicKeyBase58Check")
		return
	}
	transactorPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.TransactorPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "CreatePostAssociation: problem getting public key for the transactor")
		return
	}

	// Parse PostHashBytes from PostHashHex.
	if requestData.PostHashHex == "" {
		_AddBadRequestError(ww, "CreatePostAssociation: must provide a PostHashHex")
		return
	}
	postHashBytes, err := hex.DecodeString(requestData.PostHashHex)
	if err != nil {
		_AddBadRequestError(ww, "CreatePostAssociation: invalid PostHashHex provided")
		return
	}

	// Parse AppPublicKeyBytes from AppPublicKeyBase58Check.
	if requestData.AppPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "CreatePostAssociation: must provide an AppPublicKeyBase58Check")
		return
	}
	appPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.AppPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "CreatePostAssociation: problem getting public key for the app")
		return
	}

	// Validate AssociationType.
	if requestData.AssociationType == "" {
		_AddBadRequestError(ww, "CreatePostAssociation: must provide an AssociationType")
		return
	}

	// Validate AssociationValue.
	if requestData.AssociationValue == "" {
		_AddBadRequestError(ww, "CreatePostAssociation: must provide an AssociationValue")
		return
	}

	// Parse ExtraData.
	extraData, err := EncodeExtraDataMap(requestData.ExtraData)
	if err != nil {
		_AddBadRequestError(ww, "CreatePostAssociation: invalid ExtraData provided")
		return
	}

	// Compute the additional transaction fees as specified
	// by the request body and the node-level fees.
	additionalOutputs, err := fes.getTransactionFee(
		lib.TxnTypeCreatePostAssociation,
		transactorPublicKeyBytes,
		requestData.TransactionFees,
	)
	if err != nil {
		_AddBadRequestError(ww, "CreatePostAssociation: specified TransactionFees are invalid")
		return
	}

	// Create transaction.
	txn, totalInput, changeAmount, fees, err := fes.blockchain.CreateCreatePostAssociationTxn(
		transactorPublicKeyBytes,
		&lib.CreatePostAssociationMetadata{
			PostHash:         lib.NewBlockHash(postHashBytes),
			AppPublicKey:     lib.NewPublicKey(appPublicKeyBytes),
			AssociationType:  []byte(requestData.AssociationType),
			AssociationValue: []byte(requestData.AssociationValue),
		},
		extraData,
		requestData.MinFeeRateNanosPerKB,
		fes.backendServer.GetMempool(),
		additionalOutputs,
	)
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("CreatePostAssociation: problem creating txn: %v", err))
		return
	}

	// Construct response.
	txnBytes, err := txn.ToBytes(true)
	if err != nil {
		_AddInternalServerError(ww, "CreatePostAssociation: problem encoding txn to bytes")
		return
	}
	res := AssociationTxnResponse{
		SpendAmountNanos:  totalInput - changeAmount - fees,
		TotalInputNanos:   totalInput,
		ChangeAmountNanos: changeAmount,
		FeeNanos:          fees,
		Transaction:       txn,
		TransactionHex:    hex.EncodeToString(txnBytes),
		TxnHashHex:        txn.Hash().String(),
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, "CreatePostAssociation: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) DeletePostAssociation(ww http.ResponseWriter, req *http.Request) {
	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := DeleteAssociationRequest{}
	if err := decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "DeletePostAssociation: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "DeletePostAssociation: problem getting UTXO view")
		return
	}

	// Parse TransactorPublicKeyBytes from TransactorPublicKeyBase58Check.
	if requestData.TransactorPublicKeyBase58Check == "" {
		_AddBadRequestError(ww, "DeletePostAssociation: must provide a TransactorPublicKeyBase58Check")
		return
	}
	transactorPublicKeyBytes, _, err := fes.GetPubKeyAndProfileEntryForUsernameOrPublicKeyBase58Check(
		requestData.TransactorPublicKeyBase58Check,
		utxoView,
	)
	if err != nil {
		_AddInternalServerError(ww, "DeletePostAssociation: problem getting public key for the transactor")
		return
	}

	// Parse AssociationIDBytes from AssociationID (hex string).
	if requestData.AssociationID == "" {
		_AddBadRequestError(ww, "DeletePostAssociation: must provide an AssociationID")
		return
	}
	associationIdBytes, err := hex.DecodeString(requestData.AssociationID)
	if err != nil {
		_AddBadRequestError(ww, "DeletePostAssociation: invalid AssociationID provided")
		return
	}

	// Parse ExtraData.
	extraData, err := EncodeExtraDataMap(requestData.ExtraData)
	if err != nil {
		_AddBadRequestError(ww, "DeletePostAssociation: invalid ExtraData provided")
		return
	}

	// Compute the additional transaction fees as specified
	// by the request body and the node-level fees.
	additionalOutputs, err := fes.getTransactionFee(
		lib.TxnTypeDeletePostAssociation,
		transactorPublicKeyBytes,
		requestData.TransactionFees,
	)
	if err != nil {
		_AddBadRequestError(ww, "DeletePostAssociation: specified TransactionFees are invalid")
		return
	}

	// Create transaction.
	txn, totalInput, changeAmount, fees, err := fes.blockchain.CreateDeletePostAssociationTxn(
		transactorPublicKeyBytes,
		&lib.DeletePostAssociationMetadata{AssociationID: lib.NewBlockHash(associationIdBytes)},
		extraData,
		requestData.MinFeeRateNanosPerKB,
		fes.backendServer.GetMempool(),
		additionalOutputs,
	)
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("DeletePostAssociation: problem creating txn: %v", err))
		return
	}

	// Construct response.
	txnBytes, err := txn.ToBytes(true)
	if err != nil {
		_AddInternalServerError(ww, "DeletePostAssociation: problem encoding txn to bytes")
		return
	}
	res := AssociationTxnResponse{
		SpendAmountNanos:  totalInput - changeAmount - fees,
		TotalInputNanos:   totalInput,
		ChangeAmountNanos: changeAmount,
		FeeNanos:          fees,
		Transaction:       txn,
		TransactionHex:    hex.EncodeToString(txnBytes),
		TxnHashHex:        txn.Hash().String(),
	}
	if err = json.NewEncoder(ww).Encode(res); err != nil {
		_AddInternalServerError(ww, "DeletePostAssociation: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) GetPostAssociationByID(ww http.ResponseWriter, req *http.Request) {
	// Parse AssociationID from URL.
	vars := mux.Vars(req)
	associationIdHex, associationIdExists := vars["associationID"]
	if !associationIdExists {
		_AddBadRequestError(ww, fmt.Sprintf("GetPostAssociationByID: must provide an AssociationID"))
		return
	}

	// Parse AssociationID (BlockHash) from AssociationIdHex (string).
	associationIdBytes, err := hex.DecodeString(associationIdHex)
	if err != nil {
		_AddBadRequestError(ww, "GetPostAssociationByID: invalid AssociationID provided")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "GetPostAssociationByID: problem getting UTXO view")
		return
	}

	// Fetch AssociationEntry.
	associationEntry, err := utxoView.GetPostAssociationByID(lib.NewBlockHash(associationIdBytes))
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetPostAssociationByID: problem retrieving association: %v", err))
		return
	}
	if associationEntry == nil {
		_AddBadRequestError(ww, "GetPostAssociationByID: association not found")
		return
	}

	// Convert AssociationEntry to AssociationResponse.
	response := PostAssociationResponse{
		AssociationID:                  hex.EncodeToString(associationEntry.AssociationID.ToBytes()),
		TransactorPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TransactorPKID), false, fes.Params),
		PostHashHex:                    hex.EncodeToString(associationEntry.PostHash.ToBytes()),
		AppPublicKeyBase58Check:        lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.AppPKID), false, fes.Params),
		AssociationType:                string(associationEntry.AssociationType),
		AssociationValue:               string(associationEntry.AssociationValue),
		ExtraData:                      DecodeExtraDataMap(fes.Params, utxoView, associationEntry.ExtraData),
		BlockHeight:                    associationEntry.BlockHeight,
	}

	// JSON encode response.
	if err = json.NewEncoder(ww).Encode(response); err != nil {
		_AddInternalServerError(ww, "GetPostAssociationByID: problem encoding response as JSON")
		return
	}
}

func (fes *APIServer) GetPostAssociations(ww http.ResponseWriter, req *http.Request) {
	var err error

	// Decode request body.
	decoder := json.NewDecoder(io.LimitReader(req.Body, MaxRequestBodySizeBytes))
	requestData := PostAssociationQuery{}
	if err = decoder.Decode(&requestData); err != nil {
		_AddBadRequestError(ww, "GetPostAssociations: problem parsing request body")
		return
	}

	// Create UTXO view.
	utxoView, err := fes.backendServer.GetMempool().GetAugmentedUniversalView()
	if err != nil {
		_AddInternalServerError(ww, "GetPostAssociations: problem getting UTXO view")
		return
	}

	// Parse TransactorPKID from TransactorPublicKeyBase58Check.
	var transactorPKID *lib.PKID
	if requestData.TransactorPublicKeyBase58Check != "" {
		transactorPKID, err = fes.getPKIDFromPublicKeyBase58Check(
			utxoView, requestData.TransactorPublicKeyBase58Check,
		)
		if err != nil {
			_AddInternalServerError(ww, "GetPostAssociations: problem getting PKID for the transactor")
			return
		}
	}

	// Parse PostHash from PostHashHex.
	var postHash *lib.BlockHash
	if requestData.PostHashHex != "" {
		postHashBytes, err := hex.DecodeString(requestData.PostHashHex)
		if err != nil {
			_AddBadRequestError(ww, "GetPostAssociations: invalid PostHashHex provided")
			return
		}
		postHash = lib.NewBlockHash(postHashBytes)
	}

	// Parse AppPKID from TransactorPublicKeyBase58Check.
	var appPKID *lib.PKID
	if requestData.AppPublicKeyBase58Check != "" {
		transactorPKID, err = fes.getPKIDFromPublicKeyBase58Check(
			utxoView, requestData.AppPublicKeyBase58Check,
		)
		if err != nil {
			_AddInternalServerError(ww, "GetPostAssociations: problem getting PKID for the app")
			return
		}
	}

	// Parse LastSeenAssociationID (BlockHash) from LastSeenAssociationIdHex (string).
	var lastSeenAssociationID *lib.BlockHash
	if requestData.LastSeenAssociationID != "" {
		lastSeenAssociationIdBytes, err := hex.DecodeString(requestData.LastSeenAssociationID)
		if err != nil {
			_AddBadRequestError(ww, "GetPostAssociations: invalid LastSeenAssociationID provided")
			return
		}
		lastSeenAssociationID = lib.NewBlockHash(lastSeenAssociationIdBytes)
	}

	// Query for association entries.
	associationEntries, err := utxoView.GetPostAssociationsByAttributes(&lib.PostAssociationQuery{
		TransactorPKID:         transactorPKID,
		PostHash:               postHash,
		AppPKID:                appPKID,
		AssociationType:        []byte(requestData.AssociationType),
		AssociationTypePrefix:  []byte(requestData.AssociationTypePrefix),
		AssociationValue:       []byte(requestData.AssociationValue),
		AssociationValuePrefix: []byte(requestData.AssociationValuePrefix),
		Limit:                  requestData.Limit,
		LastSeenAssociationID:  lastSeenAssociationID,
		SortDescending:         requestData.SortDescending,
	})
	if err != nil {
		_AddInternalServerError(ww, fmt.Sprintf("GetPostAssociations: %v", err))
		return
	}

	// Convert AssociationEntries to AssociationResponses.
	associationResponses := []*PostAssociationResponse{}
	for _, associationEntry := range associationEntries {
		associationResponses = append(associationResponses, &PostAssociationResponse{
			AssociationID:                  hex.EncodeToString(associationEntry.AssociationID.ToBytes()),
			TransactorPublicKeyBase58Check: lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.TransactorPKID), false, fes.Params),
			PostHashHex:                    hex.EncodeToString(associationEntry.PostHash.ToBytes()),
			AppPublicKeyBase58Check:        lib.Base58CheckEncode(utxoView.GetPublicKeyForPKID(associationEntry.AppPKID), false, fes.Params),
			AssociationType:                string(associationEntry.AssociationType),
			AssociationValue:               string(associationEntry.AssociationValue),
			ExtraData:                      DecodeExtraDataMap(fes.Params, utxoView, associationEntry.ExtraData),
			BlockHeight:                    associationEntry.BlockHeight,
		})
	}

	// JSON encode response.
	response := PostAssociationsResponse{Associations: associationResponses}
	if err = json.NewEncoder(ww).Encode(response); err != nil {
		_AddInternalServerError(ww, "GetPostAssociations: problem encoding response as JSON")
		return
	}
}
