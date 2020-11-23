#!/usr/bin/python3

from typing import Dict, Tuple, cast

import argparse
import numpy as np
import tensorflow as tf

import flwr as fl


def main() -> None:
    """Load data, create and start FashionMnistClient."""
    parser = argparse.ArgumentParser(description="Flower")
    parser.add_argument(
        "--server_address",
        type=str,
        default="localhost:8080",
        help=f"gRPC server address (default: localhost:8080)",
    )
    parser.add_argument(
        "--source",
        type=str,
        default="local.source",
        help=f"Specifies data Source",
    )
    parser.add_argument(
        "--model",
        type=str,
        default="local.model",
        help=f"Specifies which model to use",
    )
    parser.add_argument(
        "--sink",
        type=str,
        default="local.sink",
        help=f"Specifies data Sink",
    )
    

    args = parser.parse_args()
    
    print("Source is " + args.source + "\n")
    print("Model is " + args.model + "\n")
    print("Sink is" + args.sink + "\n")



    # Build and compile Keras model
    model = tf.keras.models.Sequential(
        [
            tf.keras.layers.Flatten(input_shape=(28, 28)),
            tf.keras.layers.Dense(128, activation="relu"),
            tf.keras.layers.Dropout(0.2),
            tf.keras.layers.Dense(10, activation="softmax"),
        ]
    )
    model.compile(
        optimizer="adam", loss="sparse_categorical_crossentropy", metrics=["accuracy"]
    )

    # Implement a Flower client
    class MnistClient(fl.client.keras_client.KerasClient):
        def __init__(
            self,
            model: tf.keras.Model,
            x_train: np.ndarray,
            y_train: np.ndarray,
            x_test: np.ndarray,
            y_test: np.ndarray,
        ) -> None:
            self.model = model
            self.x_train, self.y_train = x_train, y_train
            self.x_test, self.y_test = x_test, y_test

        def get_weights(self) -> fl.common.Weights:
            return cast(fl.common.Weights, self.model.get_weights())

        def fit(
            self, weights: fl.common.Weights, config: Dict[str, str]
        ) -> Tuple[fl.common.Weights, int, int]:
            self.model.set_weights(weights)
            self.model.fit(self.x_train, self.y_train, epochs=5)
            return self.model.get_weights(), len(self.x_train), len(self.x_train)

        def evaluate(
            self, weights: fl.common.Weights, config: Dict[str, str]
        ) -> Tuple[int, float, float]:
            self.model.set_weights(weights)
            loss, accuracy = self.model.evaluate(self.x_test, self.y_test)
            return len(self.x_test), loss, accuracy

    # Load MNIST data
    (x_train, y_train), (x_test, y_test) = tf.keras.datasets.mnist.load_data()
    x_train, x_test = x_train / 255.0, x_test / 255.0

    # Instanstiate client
    client = MnistClient(model, x_train, y_train, x_test, y_test)

    # Start client
    fl.client.start_keras_client(server_address=args.server_address, client=client)


if __name__ == "__main__":
    main()
