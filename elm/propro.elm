module propro exposing (..)

import Browser
import Html exposing (Html, button, div, text, input, h3, p, br, ul, li) -- J'ai ajouté ul et li ici
import Html.Attributes exposing (placeholder, value, class)
import Html.Events exposing (onClick, onInput)
import Random
import Http
import Json.Decode as Decode


-- MAIN

main =
  Browser.element 
    { init = init
    , update = update
    , view = view
    , subscriptions = \_ -> Sub.none
    }


-- MODEL

type alias Model =
    { state : String -- Voir si la réponse est correcte
    , attempt : String -- Ce que le joueur tape
    , liste_Mots : List String -- La liste des mots possibles
    , definition : List String  -- La définition du mot secret prise depuis l'API
    , motSecret : String
    , secretVisible : Bool -- pour afficher/cacher le mot secret
    }

init : () -> ( Model, Cmd Msg ) -- Renvoie le modèle initial et la commande pour charger les mots
init _ =
    (
    { state = ""
    , attempt = ""
    , liste_Mots = []
    , definition = [] 
    , motSecret = ""
    , secretVisible = False 
    }
    ,
    Http.get
        { url = "words.txt" -- fichier des mots enregistrés, on fait une requête HTTP pour l'obtenir obligatoirement
        , expect = Http.expectString RecevoirListeMots -- Http.expectString : (Result Http.Error String -> msg) -> Http.Expect msg
        }
    )
-- Http.get : { url : String, expect : Http.Expect msg } -> Cmd msg

-- MSG

type Msg
    = UserGuess String
    | Valider
    | RecevoirListeMots (Result Http.Error String) 
    | GenererIndexAleatoire Int
    | RecevoirDefinition (Result Http.Error (List String)) -- List String pour plusieurs définitions, forme Ok/Err [String]
    | BasculerSecret 


-- DECODEUR

getDefinition : String -> Cmd Msg -- Fonction pour obtenir la définition d'un mot via une requête HTTP
getDefinition mot = 
    Http.get
        { url = "https://api.dictionaryapi.dev/api/v2/entries/en/" ++ mot
        , expect = Http.expectJson RecevoirDefinition monDecodeurDeDefinition 
        }
-- Http.expectJson : (Result Http.Error a -> msg) -> Decoder a -> Http.Expect msg ie transforme la réponse JSON en un type Http.Expect satisfaire le type de Http.get


monDecodeurDeDefinition : Decode.Decoder (List String) -- Décodeur pour extraire la liste de définitions du JSON renvoyé par l'API
monDecodeurDeDefinition =
    Decode.index 0 
        (Decode.field "meanings" 
            (Decode.list
                (Decode.field "definitions" 
                    (Decode.list
                        (Decode.field "definition" Decode.string)  -- On recupere toutes les définitions
                )
            )
        ) 
    )
    |> Decode.map List.concat -- Pour avoir une seule liste et pas des listes de listes
    -- long mais utile pour savoir ou est la casse si jamais il y a erreur dans la récupération de la définition


-- UPDATE
-- Pourquoi l'utilisation de Cmd Msg ? Car nous avons besoin de modifications par effet de bord (requêtes HTTP, génération aléatoire), parce qu'il y a interaction avec l'extérieur du programme


update : Msg -> Model -> (Model, Cmd Msg)
update msg model =
    case msg of
        UserGuess text ->
            ( { model | attempt = text }, Cmd.none ) -- On met à jour la tentative de l'utilisateur

        Valider -> --String.toLower pour ignorer la casse et String.trim pour enlever les espaces inutiles
            if String.toLower (String.trim model.attempt) == String.toLower (String.trim model.motSecret) then 
                ( { model | state = "GAGNÉ ! Bravo !", attempt = "" }, Cmd.none )
            else
                ( { model | state = "Ce n'est pas ça, réessaie...", attempt = "" }, Cmd.none )

        BasculerSecret ->
            ( { model | secretVisible = not model.secretVisible }, Cmd.none ) -- On inverse l'état d'affichage du mot secret

        RecevoirListeMots (Ok texte) -> -- (Ok texte) signifie que la requête HTTP a réussi
            let 
                liste = String.words texte -- On transforme le texte brut en liste de mots
            in
                ( { model | liste_Mots = liste } -- On met à jour la liste des mots dans model
                , Random.generate GenererIndexAleatoire (Random.int 0 (List.length liste - 1)) -- On genère l'index (le numéro) aléatoire pour choisir le mot secret
                )

        -- Obligation de gérer l'erreur si le fichier words.txt n'est pas trouvé
        RecevoirListeMots (Err _) ->
            ( { model | state = "Erreur : Impossible de charger le fichier words.txt" }, Cmd.none )

        GenererIndexAleatoire index ->
            let
                leMotChoisi = -- dans la liste de mots, on drop les premiers mots jusqu'à l'index choisi, puis on prend le premier mot restant
                    model.liste_Mots
                        |> List.drop index -- les |> passent l'argument de gauche à la fonction de droite
                        |> List.head
                        |> Maybe.withDefault "Erreur"
            in
            ( { model | 
                motSecret = leMotChoisi, -- on applique le mot choisi au modèle
                state = "Devine le mot !", -- on met à jour l'état
                secretVisible = False -- on recache le secret quand un nouveau mot est choisi
              }
            , getDefinition leMotChoisi
            )

        RecevoirDefinition (Ok def) ->
            ( { model | definition = def }, Cmd.none )

        RecevoirDefinition (Err _) ->
            ( { model | definition = ["Pas de définition trouvée pour ce mot."] }, Cmd.none )


-- VIEW
view : Model -> Html Msg
view model =
    div [ class "main-container" ] -- "box" qui contient les éléments(affichage) principaux
        [ h3 [ class "game-title" ] [ text "Jeu de Devinette" ] -- Le titre du jeu. Sa classe permet de gerer 
        
        -- MODIFICATION ICI : On passe d'un simple text à une liste (ul)
        , div [ class "definition-box" ] -- Une petite boîte dédiée à la définition.
            [ text "Définitions :" 
            , ul [] (List.map (\def -> li [] [ text def ]) model.definition) -- List.map pour afficher liste de defs
        
        , div [ class "status-message" ] -- Ici s'affiche le message (ex: "Bravo !" ou "Réessaie"). En CSS, on pourra dire que cette classe "status-message" écrit toujours en gras.
            [ text model.state ]
        
        , br [] [] -- Un simple saut de ligne (pour laisser de l'espace)
        -- La zone où le joueur tape. 
        -- "placeholder" est le texte gris qui disparaît quand on écrit.
        -- "onInput UserGuess" prévient Elm à chaque fois qu'une lettre est tapée.
        , input 
            [ class "input-field" -- On nomme l'input
            , placeholder "Tapez le mot en anglais"
            , value model.attempt
            , onInput UserGuess
            ] []
            
        , button [ class "btn-verify", onClick Valider ] [ text "Vérifier" ]-- bouton valider lance la verification une fois créer 
        
        , br [] [] -- Un simple saut de ligne (pour laisser de l'espace)

        , div [ class "cheat-section" ] 
            [ if model.secretVisible then
            -- SI le secret doit être vu : on montre le mot et le bouton pour le recacher.
                div []
                    [ p [ class "secret-label" ] [ text ("(Secret : " ++ model.motSecret ++ ")") ]
                    , button [ class "btn-cheat", onClick BasculerSecret ] [ text "Cacher l'indice" ]
                    ]
              else
              -- SINON : on montre juste le bouton pour afficher la réponse.
                button [ class "btn-cheat", onClick BasculerSecret ] [ text "Afficher la réponse (Triche)" ]
            ]
        ]]

